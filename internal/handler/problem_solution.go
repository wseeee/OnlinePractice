package handler

import (
	"OnlinePrictice/internal/helper"
	"OnlinePrictice/internal/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetProblemSolution 获取已发布题解（公共）
// @Tags 公共方法
// @Summary 获取题解
// @Param problem_identity query string true "problem_identity"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /problem-solution [get]
func GetProblemSolution(c *gin.Context) {
	problemIdentity := c.Query("problem_identity")
	if problemIdentity == "" {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "缺少 problem_identity"})
		return
	}
	s, err := model.GetPublishedSolution(problemIdentity)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 200, "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": s})
}

// UpsertProblemSolution 创建或更新题解（管理员）
// @Tags 管理员私有方法
// @Summary 创建/更新题解
// @Param Authorization header string true "Authorization"
// @Param problem_identity formData string true "problem_identity"
// @Param title formData string false "title"
// @Param content formData string true "content"
// @Param language formData string false "language"
// @Param is_published formData int false "is_published"
// @Success 200 {string} json "{"code":"200","msg":""}"
// @Router /admin/problem-solution [put]
func UpsertProblemSolution(c *gin.Context) {
	problemIdentity := c.PostForm("problem_identity")
	content := c.PostForm("content")
	if problemIdentity == "" || content == "" {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数不完整"})
		return
	}
	title := c.DefaultPostForm("title", "官方题解")
	language := c.DefaultPostForm("language", "go")
	isPublished, _ := c.GetPostForm("is_published")
	isPub := 1
	if isPublished == "0" {
		isPub = 0
	}

	// 从 JWT 获取管理员 identity
	u, _ := c.Get("user_claims")
	userIdentity := u.(*helper.UserJwt).Identity

	s := &model.ProblemSolution{
		Identity:        helper.GetUUID(),
		ProblemIdentity: problemIdentity,
		Title:           title,
		Content:         content,
		Language:        language,
		AuthorIdentity:  userIdentity,
		IsPublished:     isPub,
	}
	if err := model.UpsertSolution(s); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "保存失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "保存成功"})
}

// DeleteProblemSolution 删除题解（管理员）
// @Tags 管理员私有方法
// @Summary 删除题解
// @Param Authorization header string true "Authorization"
// @Param problem_identity query string true "problem_identity"
// @Success 200 {string} json "{"code":"200","msg":""}"
// @Router /admin/problem-solution [delete]
func DeleteProblemSolution(c *gin.Context) {
	problemIdentity := c.Query("problem_identity")
	if problemIdentity == "" {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "缺少 problem_identity"})
		return
	}
	if err := model.DeleteSolution(problemIdentity); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "删除成功"})
}
