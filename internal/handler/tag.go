package handler

import (
	"OnlinePrictice/internal/helper"
	"OnlinePrictice/internal/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ==================== 公开接口 ====================

// GetTagList 获取所有标签（公共）
func GetTagList(c *gin.Context) {
	tags, err := model.ListTags()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": tags})
}

// GetProblemTagList 获取某题目的标签（公共）
func GetProblemTagList(c *gin.Context) {
	problemIdentity := c.Query("problem_identity")
	if problemIdentity == "" {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "缺少 problem_identity"})
		return
	}
	tags, err := model.GetProblemTags(problemIdentity)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": tags})
}

// ==================== 管理员接口 ====================

// CreateTag 创建标签（管理员）
func CreateTag(c *gin.Context) {
	name := c.PostForm("name")
	color := c.DefaultPostForm("color", "#409eff")
	if name == "" {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "标签名不能为空"})
		return
	}
	tag := &model.Tag{
		Identity: helper.GetUUID(),
		Name:     name,
		Color:    color,
	}
	if err := model.CreateTag(tag); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "创建失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": tag})
}

// UpdateTag 修改标签（管理员）
func UpdateTag(c *gin.Context) {
	identity := c.PostForm("identity")
	name := c.PostForm("name")
	color := c.DefaultPostForm("color", "#409eff")
	if identity == "" || name == "" {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数不足"})
		return
	}
	if err := model.UpdateTag(identity, name, color); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "修改失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "修改成功"})
}

// DeleteTag 删除标签（管理员）
func DeleteTag(c *gin.Context) {
	identity := c.Query("identity")
	if identity == "" {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数不足"})
		return
	}
	if err := model.DeleteTag(identity); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "删除成功"})
}

// SetProblemTags 设置题目标签（管理员）
func SetProblemTags(c *gin.Context) {
	var req struct {
		ProblemIdentity string   `json:"problem_identity"`
		TagIdentities   []string `json:"tag_identities"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.ProblemIdentity == "" {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数错误"})
		return
	}
	if err := model.SetProblemTags(req.ProblemIdentity, req.TagIdentities); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "设置失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "设置成功"})
}
