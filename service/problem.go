package service

import (
	"OnlinePrictice/Models"
	"OnlinePrictice/define"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetProblemList
// @Tags 公共方法
// @Summary 问题列表
// @Param page query int false "请输入当前页，默认第一页"
// @Param size query int false "size"
// @Param keyword query string false "keyword"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /problem-list [get]
func GetProblemList(c *gin.Context) {
	size, err := strconv.Atoi(c.DefaultQuery("size", define.DefaultSize))
	if err != nil {
		log.Printf("Failed to get size: %v", err)
		return
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", define.DefaultPage))
	if err != nil {
		log.Printf("Failed to get page: %v", err)
		return
	}
	page = (page - 1) * size
	var count int64
	keyword := c.Query("keyword")

	list := make([]*Models.Problem, 0)
	tx := Models.GetProblemList(keyword)
	err = tx.Count(&count).Omit("content").Offset(page).Limit(size).Find(&list).Error
	if err != nil {
		log.Printf("Failed to get problem list: %v", err)
		return
	}
	c.JSON(200, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"list":  list,
			"count": count,
		},
	})

}
