package service

import (
	"OnlinePrictice/Models"
	"OnlinePrictice/define"
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
	page, err := strconv.Atoi(c.DefaultQuery("page", define.DefaultPage))
	if err != nil {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "参数page类型错误",
		})
		return
	}

	size, err := strconv.Atoi(c.DefaultQuery("size", define.DefaultSize))
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
	status, _ := strconv.Atoi(c.Query("status"))

	data := make([]*Models.SubmitBasic, 0)
	var count int64
	err = Models.GetSubmitList(problemidentity, useridentity, status).Count(&count).Limit(size).Offset(page).Find(&data).Error
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
