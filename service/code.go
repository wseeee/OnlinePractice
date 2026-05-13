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
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

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
	code, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Read Code Error:" + err.Error(),
		})
		return
	}
	// 代码保存
	path, err := Helper.SaveCode(code)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Code Save Error:" + err.Error(),
		})
		return
	}
	u, _ := c.Get("user_claims")
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
			"msg":  "Get Problem Error:" + err.Error(),
		})
		return
	}
	// 答案错误的channel
	WA := make(chan int)
	// 超内存的channel
	OOM := make(chan int)
	// 编译错误的channel
	CE := make(chan int)
	// 答案正确的channel
	AC := make(chan int)
	// 非法代码的channel
	EC := make(chan struct{})

	// 通过的个数
	passCount := 0
	var lock sync.Mutex
	// 提示信息
	var msg string

	// 检查代码的合法性
	v, err := Helper.CheckGoCodeValid(path)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Code Check Error:" + err.Error(),
		})
		return
	}
	if !v {
		go func() {
			EC <- struct{}{}
		}()
	} else {
		for _, testCase := range pb.TestCases {
			testCase := testCase
			go func() {
				cmd := exec.Command("go", "run", path)
				var out, stderr bytes.Buffer
				cmd.Stderr = &stderr
				cmd.Stdout = &out
				stdinPipe, err := cmd.StdinPipe()
				if err != nil {
					log.Fatalln(err)
				}
				io.WriteString(stdinPipe, testCase.Input+"\n")

				var bm runtime.MemStats
				runtime.ReadMemStats(&bm)
				if err := cmd.Run(); err != nil {
					log.Println(err, stderr.String())
					if err.Error() == "exit status 2" {
						msg = stderr.String()
						CE <- 1
						return
					}
				}
				var em runtime.MemStats
				runtime.ReadMemStats(&em)

				// 答案错误
				if testCase.Output != out.String() {
					WA <- 1
					return
				}
				// 运行超内存
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
	}

	select {
	case <-EC:
		msg = "无效代码"
		sb.Status = 6
	case <-WA:
		msg = "答案错误"
		sb.Status = 2
	case <-OOM:
		msg = "运行超内存"
		sb.Status = 4
	case <-CE:
		sb.Status = 5
	case <-AC:
		msg = "答案正确"
		sb.Status = 1
	case <-time.After(time.Millisecond * time.Duration(pb.MaxRuntime)):
		if passCount == len(pb.TestCases) {
			sb.Status = 1
			msg = "答案正确"
		} else {
			sb.Status = 3
			msg = "运行超时"
		}
	}

	if err = Models.DB.Transaction(func(tx *gorm.DB) error {
		err = tx.Create(sb).Error
		if err != nil {
			return errors.New("SubmitBasic Save Error:" + err.Error())
		}
		m := make(map[string]interface{})
		m["submit_num"] = gorm.Expr("submit_num + ?", 1)
		if sb.Status == 1 {
			m["pass_num"] = gorm.Expr("pass_num + ?", 1)
		}
		// 更新 user_basic
		err = tx.Model(new(Models.UserBasic)).Where("identity = ?", userClaim.Identity).Updates(m).Error
		if err != nil {
			return errors.New("UserBasic Modify Error:" + err.Error())
		}
		// 更新 problem_basic
		err = tx.Model(new(Models.ProblemBasic)).Where("identity = ?", problemIdentity).Updates(m).Error
		if err != nil {
			return errors.New("ProblemBasic Modify Error:" + err.Error())
		}
		return nil
	}); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Submit Error:" + err.Error(),
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
