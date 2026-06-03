package handler

import (
	"OnlinePrictice/internal/helper"
	"OnlinePrictice/internal/model"
	"OnlinePrictice/internal/sandbox"
	"OnlinePrictice/internal/service"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CodeSubmit
// @Tags 用户私有方法
// @Summary 代码提交（异步）
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
	language := c.DefaultQuery("language", "go")

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
	userClaim := u.(*helper.UserJwt)

	codePath, valid, err := service.SaveAndValidateCodeLang(code, language)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": err.Error()})
		return
	}

	submitIdentity := helper.GetUUID()

	if !valid {
		sb := &model.SubmitBasic{
			Identity:        submitIdentity,
			ProblemIdentity: problemIdentity,
			UserIdentity:    userClaim.Identity,
			Path:            codePath,
			Status:          6,
			Language:        language,
		}
		model.DB.Create(sb)
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"data": map[string]interface{}{
				"submit_identity": submitIdentity,
				"status":          6,
				"msg":             "无效代码",
			},
		})
		return
	}

	sb := &model.SubmitBasic{
		Identity:        submitIdentity,
		ProblemIdentity: problemIdentity,
		UserIdentity:    userClaim.Identity,
		Path:            codePath,
		Status:          0,
		Language:        language,
	}
	err = model.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(sb).Error; err != nil {
			return errors.New("提交记录保存失败")
		}
		m := map[string]interface{}{
			"submit_num": gorm.Expr("submit_num + ?", 1),
		}
		if err := tx.Model(new(model.UserBasic)).Where("identity = ?", userClaim.Identity).Updates(m).Error; err != nil {
			return errors.New("用户数据更新失败")
		}
		if err := tx.Model(new(model.ProblemBasic)).Where("identity = ?", problemIdentity).Updates(m).Error; err != nil {
			return errors.New("题目数据更新失败")
		}
		return nil
	})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "提交失败，请稍后重试"})
		return
	}

	task := &model.JudgeTask{
		SubmitIdentity:  submitIdentity,
		RecordType:      "submit",
		CodePath:        codePath,
		ProblemIdentity: problemIdentity,
		UserIdentity:    userClaim.Identity,
		Language:        language,
	}
	if err := model.PublishJudgeTask(model.RoutingKeySubmit, task); err != nil {
		// RabbitMQ 不可用时降级为同步判题
		passed, total, status, judgeErr := service.JudgeCodeWithLang(codePath, problemIdentity, language)
		if judgeErr != nil {
			model.DB.Model(sb).Update("status", 6)
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "判题失败: " + judgeErr.Error()})
			return
		}
		model.DB.Model(sb).Update("status", status)
		if status == 1 {
			model.DB.Model(new(model.UserBasic)).Where("identity = ?", userClaim.Identity).Update("pass_num", gorm.Expr("pass_num + ?", 1))
			model.DB.Model(new(model.ProblemBasic)).Where("identity = ?", problemIdentity).Update("pass_num", gorm.Expr("pass_num + ?", 1))
		}
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"data": map[string]interface{}{
				"submit_identity": submitIdentity,
				"status":          status,
				"passed":          passed,
				"total":           total,
				"msg":             service.StatusMsg(status),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"submit_identity": submitIdentity,
			"status":          0,
			"msg":             "已提交，排队中",
		},
	})
}

// GetLanguages 返回支持的语言列表
func GetLanguages(c *gin.Context) {
	type langInfo struct {
		Name  string `json:"name"`
		Label string `json:"label"`
	}
	labels := map[string]string{
		"go":     "Go",
		"python": "Python 3",
		"cpp":    "C++",
	}
	names := sandbox.ProfileNames()
	list := make([]langInfo, 0, len(names))
	for _, n := range names {
		label := labels[n]
		if label == "" {
			label = n
		}
		list = append(list, langInfo{Name: n, Label: label})
	}
	c.JSON(200, gin.H{"code": 200, "data": list})
}
