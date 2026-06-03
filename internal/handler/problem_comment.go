package handler

import (
	"OnlinePrictice/internal/helper"
	"OnlinePrictice/internal/model"
	"OnlinePrictice/internal/config"
	"OnlinePrictice/internal/service"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetProblemComments 获取评论列表（公共）
// @Tags 公共方法
// @Summary 评论列表
// @Param problem_identity query string true "problem_identity"
// @Param page query int false "page"
// @Param size query int false "size"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /problem-comments [get]
func GetProblemComments(c *gin.Context) {
	problemIdentity := c.Query("problem_identity")
	if problemIdentity == "" {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "缺少 problem_identity"})
		return
	}
	size, _ := strconv.Atoi(c.DefaultQuery("size", config.DefaultSize))
	page, _ := strconv.Atoi(c.DefaultQuery("page", config.DefaultPage))

	comments, count, err := model.ListTopLevelComments(problemIdentity, page, size)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "查询失败"})
		return
	}

	// 批量查用户名
	fillUserNames(comments)

	// 批量查回复
	if len(comments) > 0 {
		parentIds := make([]string, len(comments))
		for i, cm := range comments {
			parentIds[i] = cm.Identity
		}
		replies, _ := model.ListRepliesByParentIdentities(parentIds)
		fillUserNames(replies)
		// 归类回复
		replyMap := make(map[string][]*model.ProblemComment)
		for _, r := range replies {
			replyMap[r.ParentIdentity] = append(replyMap[r.ParentIdentity], r)
		}
		for _, cm := range comments {
			cm.Replies = replyMap[cm.Identity]
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"list":  comments,
			"count": count,
		},
	})
}

// PostProblemComment 发表评论/回复（用户）
// @Tags 用户私有方法
// @Summary 发表评论
// @Param Authorization header string true "Authorization"
// @Param body body object true "body"
// @Success 200 {string} json "{"code":"200","msg":""}"
// @Router /user/problem-comment [post]
func PostProblemComment(c *gin.Context) {
	var req struct {
		ProblemIdentity string `json:"problem_identity"`
		ParentIdentity  string `json:"parent_identity"`
		Content         string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数错误"})
		return
	}
	req.Content = strings.TrimSpace(req.Content)
	if req.ProblemIdentity == "" || req.Content == "" {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "内容不能为空"})
		return
	}
	if len(req.Content) > 2000 {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "评论最长 2000 字符"})
		return
	}

	// 校验回复深度：仅允许一级回复
	var parentComment *model.ProblemComment
	if req.ParentIdentity != "" {
		pc, err := model.GetCommentByIdentity(req.ParentIdentity)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "父评论不存在"})
			return
		}
		if pc.ParentIdentity != "" {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "仅支持一级回复"})
			return
		}
		parentComment = pc
	}

	u, _ := c.Get("user_claims")
	userIdentity := u.(*helper.UserJwt).Identity
	comment := &model.ProblemComment{
		Identity:        helper.GetUUID(),
		ProblemIdentity: req.ProblemIdentity,
		UserIdentity:    userIdentity,
		ParentIdentity:  req.ParentIdentity,
		Content:         req.Content,
		Status:          1,
	}
	if err := model.CreateComment(comment); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "发表失败"})
		return
	}

	// 回复时发送通知给父评论作者
	if parentComment != nil && parentComment.UserIdentity != "" && parentComment.UserIdentity != userIdentity {
		var problem model.ProblemBasic
		model.DB.Where("identity = ?", req.ProblemIdentity).First(&problem)
		service.SendNotification(parentComment.UserIdentity, "comment_reply",
			"有人回复了你的评论",
			"在题目「"+problem.Title+"」中收到了新回复："+truncateStr(req.Content, 100),
			req.ProblemIdentity)
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "发表成功"})
}

func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// DeleteProblemComment 删除自己的评论（用户）
// @Tags 用户私有方法
// @Summary 删除评论
// @Param Authorization header string true "Authorization"
// @Param identity query string true "identity"
// @Success 200 {string} json "{"code":"200","msg":""}"
// @Router /user/problem-comment [delete]
func DeleteProblemComment(c *gin.Context) {
	identity := c.Query("identity")
	if identity == "" {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "缺少 identity"})
		return
	}
	comment, err := model.GetCommentByIdentity(identity)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "评论不存在"})
		return
	}
	u, _ := c.Get("user_claims")
	userIdentity := u.(*helper.UserJwt).Identity
	if comment.UserIdentity != userIdentity {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "无权删除"})
		return
	}
	if err := model.SoftDeleteComment(identity); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "删除成功"})
}

// AdminHideProblemComment 管理员隐藏评论
// @Tags 管理员私有方法
// @Summary 隐藏评论
// @Param Authorization header string true "Authorization"
// @Param identity query string true "identity"
// @Success 200 {string} json "{"code":"200","msg":""}"
// @Router /admin/problem-comment [delete]
func AdminHideProblemComment(c *gin.Context) {
	identity := c.Query("identity")
	if identity == "" {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "缺少 identity"})
		return
	}
	if err := model.AdminHideComment(identity); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "操作失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "已隐藏"})
}

// fillUserNames 批量填充评论的用户名
func fillUserNames(comments []*model.ProblemComment) {
	if len(comments) == 0 {
		return
	}
	ids := make([]string, 0, len(comments))
	for _, cm := range comments {
		ids = append(ids, cm.UserIdentity)
	}
	var users []model.UserBasic
	model.DB.Where("identity IN ?", ids).Find(&users)
	nameMap := make(map[string]string, len(users))
	for _, u := range users {
		nameMap[u.Identity] = u.Name
	}
	for _, cm := range comments {
		cm.UserName = nameMap[cm.UserIdentity]
	}
}
