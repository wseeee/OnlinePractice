package handler

import (
	"OnlinePrictice/internal/model"
	"OnlinePrictice/internal/config"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetSubmitList
// @Tags 公共方法
// @Summary 问题列表
// @Param page query int false "请输入当前页，默认第一页"
// @Param size query int false "size"
// @Param problem_identity query string false "problem_identity"
// @Param user_identity query string false "user_identity "
// @Param status query int false "status"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /submit-list [get]
func GetSubmitList(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", config.DefaultPage))
	if err != nil {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "参数page类型错误",
		})
		return
	}

	size, err := strconv.Atoi(c.DefaultQuery("size", config.DefaultSize))
	if err != nil {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "参数size类型错误",
		})
		return
	}
	page = (page - 1) * size

	problemidentity := c.Query("problem_identity")
	useridentity := c.Query("user_identity")
	status := -1
	if s := c.Query("status"); s != "" {
		status, _ = strconv.Atoi(s)
	}

	data := make([]*model.SubmitBasic, 0)
	var count int64
	err = model.GetSubmitList(problemidentity, useridentity, status).Count(&count).Limit(size).Offset(page).Find(&data).Error
	if err != nil {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "查询失败",
		})
		return
	}
	c.JSON(200, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"count": count,
			"list":  data,
		},
	})
}
