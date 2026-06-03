package handler

import (
	"OnlinePrictice/internal/helper"
	"OnlinePrictice/internal/model"
	"OnlinePrictice/internal/config"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetProblemCodeShares 获取某题的公开代码列表（公共）
// @Tags 公共方法
// @Summary 代码分享列表
// @Param problem_identity query string true "problem_identity"
// @Param page query int false "page"
// @Param size query int false "size"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /problem-code-shares [get]
func GetProblemCodeShares(c *gin.Context) {
	problemIdentity := c.Query("problem_identity")
	if problemIdentity == "" {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "缺少 problem_identity"})
		return
	}
	size, _ := strconv.Atoi(c.DefaultQuery("size", config.DefaultSize))
	page, _ := strconv.Atoi(c.DefaultQuery("page", config.DefaultPage))

	shares, count, err := model.ListSharesByProblem(problemIdentity, page, size)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "查询失败"})
		return
	}

	// 填充用户名
	type shareItem struct {
		Identity       string `json:"identity"`
		Title          string `json:"title"`
		UserName       string `json:"user_name"`
		SubmitIdentity string `json:"submit_identity"`
		ViewCount      int    `json:"view_count"`
		CreatedAt      string `json:"created_at"`
	}
	items := make([]shareItem, len(shares))
	userIds := make([]string, len(shares))
	for i, s := range shares {
		userIds[i] = s.UserIdentity
	}
	var users []model.UserBasic
	model.DB.Where("identity IN ?", userIds).Find(&users)
	nameMap := make(map[string]string, len(users))
	for _, u := range users {
		nameMap[u.Identity] = u.Name
	}
	for i, s := range shares {
		items[i] = shareItem{
			Identity:       s.Identity,
			Title:          s.Title,
			UserName:       nameMap[s.UserIdentity],
			SubmitIdentity: s.SubmitIdentity,
			ViewCount:      s.ViewCount,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"list":  items,
			"count": count,
		},
	})
}

// GetCodeShareDetail 获取分享详情 + 代码正文（公共）
// @Tags 公共方法
// @Summary 代码分享详情
// @Param identity query string true "identity"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /code-share-detail [get]
func GetCodeShareDetail(c *gin.Context) {
	identity := c.Query("identity")
	if identity == "" {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "缺少 identity"})
		return
	}
	share, err := model.GetShareByIdentity(identity)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "分享不存在"})
		return
	}

	// 查关联提交获取代码路径
	var submit model.SubmitBasic
	if err := model.DB.Where("identity = ?", share.SubmitIdentity).First(&submit).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "提交记录不存在"})
		return
	}

	// 读取代码文件，限制 64KB
	code := ""
	if submit.Path != "" {
		data, err := os.ReadFile(submit.Path)
		if err == nil && len(data) <= 64*1024 {
			code = string(data)
		} else if len(data) > 64*1024 {
			code = string(data[:64*1024])
		}
	}

	// 浏览量 +1
	model.DB.Model(new(model.CodeShare)).Where("identity = ?", identity).Update("view_count", model.DB.Raw("view_count + 1"))

	// 查用户名
	var user model.UserBasic
	model.DB.Where("identity = ?", share.UserIdentity).First(&user)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"identity":   share.Identity,
			"title":      share.Title,
			"code":       code,
			"language":   "go",
			"user_name":  user.Name,
			"created_at": share.CreatedAt,
		},
	})
}

// PostCodeShare 公开自己的 AC 代码（用户）
// @Tags 用户私有方法
// @Summary 公开代码
// @Param Authorization header string true "Authorization"
// @Param body body object true "body"
// @Success 200 {string} json "{"code":"200","msg":""}"
// @Router /user/code-share [post]
func PostCodeShare(c *gin.Context) {
	var req struct {
		SubmitIdentity string `json:"submit_identity"`
		Title          string `json:"title"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数错误"})
		return
	}
	if req.SubmitIdentity == "" {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "缺少 submit_identity"})
		return
	}

	// 查提交记录
	var submit model.SubmitBasic
	if err := model.DB.Where("identity = ?", req.SubmitIdentity).First(&submit).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "提交不存在"})
		return
	}

	// 校验归属
	u, _ := c.Get("user_claims")
	userIdentity := u.(*helper.UserJwt).Identity
	if submit.UserIdentity != userIdentity {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "只能分享自己的提交"})
		return
	}

	// 校验 AC
	if submit.Status != 1 {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "仅 AC 提交可分享"})
		return
	}

	// 校验未重复分享
	if model.ExistsShareBySubmit(req.SubmitIdentity) {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "该提交已分享"})
		return
	}

	title := req.Title
	if title == "" {
		title = "AC 代码"
	}
	share := &model.CodeShare{
		Identity:        helper.GetUUID(),
		SubmitIdentity:  req.SubmitIdentity,
		ProblemIdentity: submit.ProblemIdentity,
		UserIdentity:    submit.UserIdentity,
		Title:           title,
		ViewCount:       0,
	}
	if err := model.CreateShare(share); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "分享失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "分享成功"})
}

// DeleteCodeShare 取消分享（用户）
// @Tags 用户私有方法
// @Summary 取消分享
// @Param Authorization header string true "Authorization"
// @Param identity query string true "identity"
// @Success 200 {string} json "{"code":"200","msg":""}"
// @Router /user/code-share [delete]
func DeleteCodeShare(c *gin.Context) {
	identity := c.Query("identity")
	if identity == "" {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "缺少 identity"})
		return
	}
	share, err := model.GetShareByIdentity(identity)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "分享不存在"})
		return
	}
	u, _ := c.Get("user_claims")
	userIdentity := u.(*helper.UserJwt).Identity
	if share.UserIdentity != userIdentity {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "无权操作"})
		return
	}
	if err := model.DeleteShareByIdentity(identity); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "取消失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "已取消分享"})
}
