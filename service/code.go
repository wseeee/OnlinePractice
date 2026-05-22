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

// JudgeCode 判题公共函数，返回通过数、总用例数、状态码、代码保存路径、错误
func JudgeCode(code []byte, problemIdentity string) (passed, total int, status int, path string, err error) {
	// 代码保存
	path, err = Helper.SaveCode(code)
	if err != nil {
		return 0, 0, 0, "", err
	}

	// 查询题目及测试用例
	pb := new(Models.ProblemBasic)
	err = Models.DB.Where("identity = ?", problemIdentity).Preload("TestCases").First(pb).Error
	if err != nil {
		return 0, 0, 0, path, err
	}
	if len(pb.TestCases) == 0 {
		return 0, 0, 0, path, errors.New("该题目没有测试用例")
	}
	total = len(pb.TestCases)

	// 代码合法性校验
	v, err := Helper.CheckGoCodeValid(path)
	if err != nil {
		return 0, total, 0, path, err
	}
	if !v {
		return 0, total, 6, path, nil // 无效代码
	}

	// 并发执行测试用例
	WA := make(chan int, total)
	OOM := make(chan int, total)
	CE := make(chan string, total)
	AC := make(chan int, total)

	passCount := 0
	var lock sync.Mutex

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

			actual := strings.TrimSpace(out.String())
			expected := strings.TrimSpace(testCase.Output)
			if expected != actual {
				WA <- 1
				return
			}
			if em.Alloc/1024-(bm.Alloc/1024) > uint64(pb.MaxMem) {
				OOM <- 1
				return
			}
			lock.Lock()
			passCount++
			if passCount == total {
				AC <- 1
			}
			lock.Unlock()
		}()
	}

	go func() {
		wg.Wait()
		close(WA)
		close(OOM)
		close(CE)
		close(AC)
	}()

	select {
	case <-WA:
		lock.Lock()
		return passCount, total, 2, path, nil // WA
	case <-OOM:
		lock.Lock()
		return passCount, total, 4, path, nil // MLE
	case errMsg := <-CE:
		lock.Lock()
		_ = errMsg
		return passCount, total, 5, path, nil // CE
	case <-AC:
		lock.Lock()
		return total, total, 1, path, nil // AC
	case <-time.After(time.Millisecond * time.Duration(pb.MaxRuntime)):
		lock.Lock()
		if passCount == total {
			return total, total, 1, path, nil
		}
		return passCount, total, 3, path, nil // TLE
	}
}

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
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数不完整"})
		return
	}

	code, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "读取代码失败"})
		return
	}

	u, exists := c.Get("user_claims")
	if !exists {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "用户信息获取失败"})
		return
	}
	userClaim := u.(*Helper.UserJwt)

	_, _, status, _, err := JudgeCode(code, problemIdentity)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": err.Error()})
		return
	}

	sb := &Models.SubmitBasic{
		Identity:        Helper.GetUUID(),
		ProblemIdentity: problemIdentity,
		UserIdentity:    userClaim.Identity,
		Status:          status,
	}

	msgMap := map[int]string{1: "答案正确", 2: "答案错误", 3: "运行超时", 4: "运行超内存", 5: "编译错误", 6: "无效代码"}
	msg := msgMap[status]

	saveSubmitAndRespond(c, sb, userClaim, problemIdentity, msg)
}

func saveSubmitAndRespond(c *gin.Context, sb *Models.SubmitBasic, userClaim *Helper.UserJwt, problemIdentity, msg string) {
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
