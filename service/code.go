package service

import (
	"OnlinePrictice/Helper"
	"OnlinePrictice/Models"
	"bytes"
	"errors"
	"io"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const maxOutputSize = 1024 * 1024 // 1MB 输出限制

// CodeSubmit
// @Tags 用户私有方法
// @Summary 代码提交
// @Param Authorization header string true "Authorization"
// @Param problem_identity query string true "problem_identity"
// @Param code body string true "code"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /user/code-submit [post]
func CodeSubmit(c *gin.Context) {
	problemIdentity := c.Query("problem_identity")
	if problemIdentity == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "参数不完整",
		})
		return
	}

	code, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "读取代码失败",
		})
		return
	}
	// 代码保存
	path, err := Helper.SaveCode(code)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "代码保存失败",
		})
		return
	}
	u, exists := c.Get("user_claims")
	if !exists {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "用户信息获取失败",
		})
		return
	}
	userClaim := u.(*Helper.UserJwt)
	sb := &Models.SubmitBasic{
		Identity:        Helper.GetUUID(),
		ProblemIdentity: problemIdentity,
		UserIdentity:    userClaim.Identity,
		Path:            path,
	}
	// 代码判断
	pb := new(Models.ProblemBasic)
	err = Models.DB.Where("identity = ?", problemIdentity).Preload("TestCases").First(pb).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "题目查询失败",
		})
		return
	}

	if len(pb.TestCases) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "该题目没有测试用例",
		})
		return
	}

	// 检查代码的合法性
	v, err := Helper.CheckGoCodeValid(path)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "代码检查失败",
		})
		return
	}
	if !v {
		sb.Status = 6
		saveSubmitAndRespond(c, sb, userClaim, problemIdentity, "无效代码", pb)
		return
	}

	// 使用缓冲 channel 防止 goroutine 泄漏
	WA := make(chan int, len(pb.TestCases))
	OOM := make(chan int, len(pb.TestCases))
	CE := make(chan string, len(pb.TestCases))
	AC := make(chan int, len(pb.TestCases))

	passCount := 0
	var lock sync.Mutex
	var msg string
	var msgLock sync.Mutex

	var wg sync.WaitGroup

	for _, testCase := range pb.TestCases {
		testCase := testCase
		wg.Add(1)
		go func() {
			defer wg.Done()
			cmd := exec.Command("go", "run", path)
			var out, stderr bytes.Buffer
			cmd.Stderr = &stderr
			cmd.Stdout = &out
			stdinPipe, err := cmd.StdinPipe()
			if err != nil {
				log.Printf("StdinPipe error: %v", err)
				CE <- "内部错误"
				return
			}
			io.WriteString(stdinPipe, testCase.Input+"\n")
			stdinPipe.Close()

			var bm runtime.MemStats
			runtime.ReadMemStats(&bm)
			if err := cmd.Run(); err != nil {
				errMsg := stderr.String()
				if len(errMsg) > 500 {
					errMsg = errMsg[:500]
				}
				if strings.Contains(err.Error(), "exit status 2") || strings.Contains(errMsg, "syntax error") {
					CE <- errMsg
					return
				}
			}
			var em runtime.MemStats
			runtime.ReadMemStats(&em)

			// 答案比较（去除首尾空白）
			actual := strings.TrimSpace(out.String())
			expected := strings.TrimSpace(testCase.Output)
			if expected != actual {
				WA <- 1
				return
			}
			// 运行超内存 (注意: 此检测方式仅测 Go runtime 内存，不测子进程)
			if em.Alloc/1024-(bm.Alloc/1024) > uint64(pb.MaxMem) {
				OOM <- 1
				return
			}
			lock.Lock()
			passCount++
			if passCount == len(pb.TestCases) {
				AC <- 1
			}
			lock.Unlock()
		}()
	}

	// 等待所有 goroutine 完成后关闭 channel，防止泄漏
	go func() {
		wg.Wait()
		close(WA)
		close(OOM)
		close(CE)
		close(AC)
	}()

	// 等待第一个结果或超时
	select {
	case <-WA:
		msgLock.Lock()
		msg = "答案错误"
		msgLock.Unlock()
		sb.Status = 2
	case <-OOM:
		msgLock.Lock()
		msg = "运行超内存"
		msgLock.Unlock()
		sb.Status = 4
	case errMsg := <-CE:
		msgLock.Lock()
		msg = errMsg
		msgLock.Unlock()
		sb.Status = 5
	case <-AC:
		msgLock.Lock()
		msg = "答案正确"
		msgLock.Unlock()
		sb.Status = 1
	case <-time.After(time.Millisecond * time.Duration(pb.MaxRuntime)):
		lock.Lock()
		if passCount == len(pb.TestCases) {
			sb.Status = 1
			msg = "答案正确"
		} else {
			sb.Status = 3
			msg = "运行超时"
		}
		lock.Unlock()
	}

	saveSubmitAndRespond(c, sb, userClaim, problemIdentity, msg, pb)
}

func saveSubmitAndRespond(c *gin.Context, sb *Models.SubmitBasic, userClaim *Helper.UserJwt, problemIdentity, msg string, pb *Models.ProblemBasic) {
	if err := Models.DB.Transaction(func(tx *gorm.DB) error {
		err := tx.Create(sb).Error
		if err != nil {
			return errors.New("提交记录保存失败")
		}
		m := make(map[string]interface{})
		m["submit_num"] = gorm.Expr("submit_num + ?", 1)
		if sb.Status == 1 {
			m["pass_num"] = gorm.Expr("pass_num + ?", 1)
		}
		// 更新 user_basic
		err = tx.Model(new(Models.UserBasic)).Where("identity = ?", userClaim.Identity).Updates(m).Error
		if err != nil {
			return errors.New("用户数据更新失败")
		}
		// 更新 problem_basic
		err = tx.Model(new(Models.ProblemBasic)).Where("identity = ?", problemIdentity).Updates(m).Error
		if err != nil {
			return errors.New("题目数据更新失败")
		}
		return nil
	}); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "提交失败，请稍后重试",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"status": sb.Status,
			"msg":    msg,
		},
	})
}
