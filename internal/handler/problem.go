package handler

import (
	"OnlinePrictice/internal/helper"
	"OnlinePrictice/internal/model"
	"OnlinePrictice/internal/config"
	"OnlinePrictice/internal/service"
	"errors"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetProblemList
// @Tags 公共方法
// @Summary 问题列表
// @Param page query int false "请输入当前页，默认第一页"
// @Param size query int false "size"
// @Param keyword query string false "keyword"
// @Param category_identity query string false "category_identity "
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /problem-list [get]
func GetProblemList(c *gin.Context) {
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
	categoryIdentity := c.Query("category_identity")
	difficulty, _ := strconv.Atoi(c.Query("difficulty"))

	list := make([]*model.ProblemBasic, 0)
	tx := model.GetProblemList(keyword, categoryIdentity, difficulty)
	err = tx.Count(&count).Omit("content").Offset(page).Limit(size).Find(&list).Error
	if err != nil {
		log.Printf("Failed to get problem list: %v", err)
		c.JSON(200, gin.H{"code": -1, "msg": "查询失败: " + err.Error()})
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

// GetProblemDetail
// @Tags 公共方法
// @Summary 问题详情
// @Param identity query string false "problemidentity "
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /problem-detail [get]
func GetProblemDetail(c *gin.Context) {
	identity := c.Query("identity")
	if identity == "" {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "参数不完整",
		})
		return
	}
	var data *model.ProblemBasic
	err := model.GetProblemDetail(identity).Find(&data).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(200, gin.H{
				"code": -1,
				"msg":  "题目不存在",
			})
			return
		}
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "查询失败",
		})
		return
	}
	c.JSON(200, gin.H{
		"code": 200,
		"data": data,
	})
}

// CreateProblem
// @Tags 管理员私有方法
// @Summary 创建问题
// @Param Authorization header string true "Authorization"
// @Param title formData string true "title"
// @Param content formData string true "content"
// @Param category_ids formData array false "category_ids"
// @Param test_cases formData array false "test_cases"
// @Param max_mem formData int false "max_mem"
// @Param max_runtime formData int false  "max_runtime"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /admin/problem-create [post]
func CreateProblem(c *gin.Context) {
	title := c.PostForm("title")
	content := c.PostForm("content")
	maxRuntime, _ := strconv.Atoi(c.PostForm("max_runtime"))
	maxMem, _ := strconv.Atoi(c.PostForm("max_mem"))
	categoryIds := c.PostFormArray("category_ids")
	testCases := c.PostFormArray("test_cases")
	difficultyMode, _ := strconv.Atoi(c.PostForm("difficulty_mode"))
	difficultyVal, _ := strconv.Atoi(c.PostForm("difficulty"))
	if title == "" || content == "" || len(categoryIds) == 0 || len(testCases) == 0 || maxRuntime == 0 || maxMem == 0 {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "参数不齐全",
		})
		return
	}
	identity := helper.GetUUID()
	data := &model.ProblemBasic{
		Identity:       identity,
		Title:          title,
		Content:        content,
		MaxMem:         maxMem,
		MaxRuntime:     maxRuntime,
		DifficultyMode: difficultyMode,
		Difficulty:     difficultyVal,
		PassRate:       -1,
	}
	//处理分类
	categoryBasics := make([]*model.ProblemCategory, 0)
	for _, id := range categoryIds {
		atoi, _ := strconv.Atoi(id)
		categoryBasics = append(categoryBasics, &model.ProblemCategory{
			CategoryId: uint(atoi),
			ProblemId:  data.ID,
		})
	}
	data.ProblemCategories = categoryBasics
	//处理测试用例
	testCaseBasics, err := service.ParseTestCases(testCases, identity)
	if err != nil {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  err.Error(),
		})
		return
	}
	data.TestCases = testCaseBasics

	err = model.CreateProblem(data).Error
	if err != nil {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "创建失败",
		})
		return
	}
	c.JSON(200, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"identity": data.Identity,
		},
	})
}

// UpdateProblem
// @Tags 管理员私有方法
// @Summary 问题修改
// @Param Authorization header string true "Authorization"
// @Param identity  query string true "identity"
// @Param title formData string true "title"
// @Param content formData string true "content"
// @Param max_mem formData int false "max_mem"
// @Param max_runtime formData int false  "max_runtime"
// @Param category_id formData array false "category_id"
// @Param test_cases formData array false "test_cases"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /admin/problem-update [put]
func UpdateProblem(c *gin.Context) {
	identity := c.Query("identity")
	title := c.PostForm("title")
	content := c.PostForm("content")
	maxRuntime, _ := strconv.Atoi(c.PostForm("max_runtime"))
	maxMem, _ := strconv.Atoi(c.PostForm("max_mem"))
	categoryIds := c.PostFormArray("category_id")
	testCases := c.PostFormArray("test_cases")
	difficultyMode, _ := strconv.Atoi(c.PostForm("difficulty_mode"))
	difficultyVal, _ := strconv.Atoi(c.PostForm("difficulty"))
	if identity == "" || title == "" || content == "" || len(categoryIds) == 0 || len(testCases) == 0 || maxRuntime == 0 || maxMem == 0 {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "参数不齐全",
		})
		return
	}

	// 先解析测试用例，在事务外返回参数错误
	testCaseBasics, err := service.ParseTestCases(testCases, identity)
	if err != nil {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  err.Error(),
		})
		return
	}

	err = model.DB.Transaction(func(tx *gorm.DB) error {
		//问题基础信息的保存problem_basic
		problemBasic := &model.ProblemBasic{
			Identity:   identity,
			Title:      title,
			Content:    content,
			MaxMem:     maxMem,
			MaxRuntime: maxRuntime,
		}
		// 难度字段
		if difficultyMode > 0 {
			problemBasic.DifficultyMode = difficultyMode
		}
		if difficultyVal > 0 {
			problemBasic.Difficulty = difficultyVal
		}
		err := tx.Model(new(model.ProblemBasic)).
			Where("identity = ?", identity).
			Updates(problemBasic).Error
		if err != nil {
			return errors.New("题目信息更新失败")
		}
		//查询问题详情
		err = tx.Where("identity = ?", identity).First(&problemBasic).Error
		if err != nil {
			return errors.New("题目查询失败")
		}
		//关联问题分类的更新
		err = tx.Where("problem_id = ?", problemBasic.ID).Delete(&model.ProblemCategory{}).Error
		if err != nil {
			return errors.New("分类关联删除失败")
		}
		categoryBasics := make([]*model.ProblemCategory, 0)
		for _, id := range categoryIds {
			atoi, _ := strconv.Atoi(id)
			categoryBasics = append(categoryBasics, &model.ProblemCategory{
				CategoryId: uint(atoi),
				ProblemId:  problemBasic.ID,
			})
		}
		err = tx.Create(&categoryBasics).Error
		if err != nil {
			return errors.New("分类关联创建失败")
		}
		//关联测试用例的更新
		err = tx.Where("problem_identity = ?", identity).Delete(&model.TestCase{}).Error
		if err != nil {
			return errors.New("测试用例删除失败")
		}

		err = tx.Create(&testCaseBasics).Error
		if err != nil {
			return errors.New("测试用例创建失败")
		}

		return nil
	})
	if err != nil {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "更新失败，请稍后重试",
		})
		return
	}
	// 自动模式下重算难度
	if difficultyMode == 1 || difficultyMode == 0 {
		_ = service.RecalculateProblemDifficulty(identity)
	}
	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "修改成功",
	})
}
