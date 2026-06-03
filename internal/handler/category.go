package handler

import (
	"OnlinePrictice/internal/helper"
	"OnlinePrictice/internal/model"
	"OnlinePrictice/internal/config"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetCategoryList
// @Tags 管理员私有方法
// @Summary 分类列表
// @Param Authorization header string true "Authorization"
// @Param page query int false "请输入当前页，默认第一页"
// @Param size query int false "size"
// @Param keyword query string false "keyword"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /admin/category-list [get]
func GetCategoryList(c *gin.Context) {
	size, err := strconv.Atoi(c.DefaultQuery("size", config.DefaultSize))
	if err != nil {
		c.JSON(200, gin.H{"code": -1, "msg": "参数 size 类型错误"})
		return
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", config.DefaultPage))
	if err != nil {
		c.JSON(200, gin.H{"code": -1, "msg": "参数 page 类型错误"})
		return
	}
	page = (page - 1) * size

	var count int64
	keyword := c.Query("keyword")
	categorylist := make([]*model.CategoryBasic, 0)
	err = model.GetCategoryList(keyword).
		Count(&count).Offset(page).Limit(size).
		Find(&categorylist).Error
	if err != nil {
		log.Printf("Failed to get category list: %v", err)
		c.JSON(200, gin.H{"code": -1, "msg": "查询失败"})
		return
	}
	c.JSON(200, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"list":  categorylist,
			"count": count,
		},
	})
}

// CreateCategory
// @Tags 管理员私有方法
// @Summary 分类创建
// @Param Authorization header string true "Authorization"
// @Param name formData string true "name"
// @Param parent_id formData int false "parent_id"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /admin/category-create [post]
func CreateCategory(c *gin.Context) {
	name := c.PostForm("name")
	parent_id, _ := strconv.Atoi(c.PostForm("parent_id"))

	if name == "" {
		c.JSON(200, gin.H{"code": -1, "msg": "分类名称不能为空"})
		return
	}

	category := &model.CategoryBasic{
		Identity: helper.GetUUID(),
		Name:     name,
		ParentId: parent_id,
	}
	err := model.CreateCategory(category).Error
	if err != nil {
		log.Printf("Failed to create category: %v", err)
		c.JSON(200, gin.H{"code": -1, "msg": "创建失败"})
		return
	}
	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "创建成功",
	})
}

// UpdateCategory
// @Tags 管理员私有方法
// @Summary 分类修改
// @Param Authorization header string true "Authorization"
// @Param identity formData string true "identity"
// @Param name formData string true "name"
// @Param parent_id formData int false "parent_id"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /admin/category-update [put]
func UpdateCategory(c *gin.Context) {
	identity := c.PostForm("identity")
	name := c.PostForm("name")
	parent_id, _ := strconv.Atoi(c.PostForm("parent_id"))
	if identity == "" || name == "" {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "参数不足",
		})
		return
	}

	category := &model.CategoryBasic{
		Identity: identity,
		Name:     name,
		ParentId: parent_id,
	}
	err := model.UpdateCategory(identity, category).Error
	if err != nil {
		log.Printf("Failed to update category: %v", err)
		c.JSON(200, gin.H{"code": -1, "msg": "修改失败"})
		return
	}
	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "修改成功",
	})
}

// DeleteCategory
// @Tags 管理员私有方法
// @Summary 分类删除
// @Param Authorization header string true "Authorization"
// @Param identity  query string true "identity"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /admin/category-delete [delete]
func DeleteCategory(c *gin.Context) {
	identity := c.Query("identity")
	if identity == "" {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "参数不足",
		})
		return
	}
	var count int64
	err := model.DB.Model(&model.ProblemCategory{}).
		Where("category_id = (SELECT id FROM category_basic WHERE identity = ? LIMIT 1)", identity).
		Count(&count).Error
	if err != nil {
		c.JSON(200, gin.H{"code": -1, "msg": "查询失败"})
		return
	}
	if count > 0 {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "该分类下有题目，请先删除题目",
		})
		return
	}
	err = model.DeleteCategory(identity).Error
	if err != nil {
		log.Printf("Failed to delete category: %v", err)
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "删除失败",
		})
		return
	}
	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "删除成功",
	})
}
