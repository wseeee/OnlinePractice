package service

import (
	"OnlinePrictice/Helper"
	"OnlinePrictice/Models"
	"OnlinePrictice/define"
	"encoding/json"
	"fmt"
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
	categoryIdentity := c.Query("category_identity")

	list := make([]*Models.ProblemBasic, 0)
	tx := Models.GetProblemList(keyword, categoryIdentity)
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
	var data *Models.ProblemBasic
	err := Models.GetProblemDetail(identity).Find(&data).Error
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
			"msg":  "系统错误",
		})
		return
	}
	c.JSON(200, gin.H{
		"code": 200,
		"data": data,
	})
}

// GreateProblem
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
func GreateProblem(c *gin.Context) {
	title := c.PostForm("title")
	content := c.PostForm("content")
	maxRuntime, _ := strconv.Atoi(c.PostForm("max_runtime"))
	maxMem, _ := strconv.Atoi(c.PostForm("max_mem"))
	categoryIds := c.PostFormArray("category_ids")
	testCases := c.PostFormArray("test_cases")
	//{"input":"3 4\n","output":"7\n"}
	if title == "" || content == "" || len(categoryIds) == 0 || len(testCases) == 0 || maxRuntime == 0 || maxMem == 0 {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "参数不齐全",
		})
		return
	}
	identity := Helper.GetUUID()
	data := &Models.ProblemBasic{
		Identity:   identity,
		Title:      title,
		Content:    content,
		MaxMem:     maxMem,
		MaxRuntime: maxRuntime,
	}
	//处理分类
	categoryBasics := make([]*Models.ProblemCategory, 0)
	for _, id := range categoryIds {
		atoi, _ := strconv.Atoi(id)
		categoryBasics = append(categoryBasics, &Models.ProblemCategory{
			CategoryId: uint(atoi),
			ProblemId:  data.ID,
		})
	}
	data.ProblemCategories = categoryBasics
	//处理测试用例
	testCaseBasics := make([]*Models.TestCase, 0)
	for _, testCase := range testCases {
		caseMap := make(map[string]string)
		err := json.Unmarshal([]byte(testCase), &caseMap)
		if err != nil {
			c.JSON(200, gin.H{
				"code": -1,
				"msg":  "参数错误",
			})
			return
		}
		if _, ok := caseMap["input"]; !ok {
			c.JSON(200, gin.H{
				"code": -1,
				"msg":  "输入错误",
			})
			return
		}
		if _, ok := caseMap["output"]; !ok {
			c.JSON(200, gin.H{
				"code": -1,
				"msg":  "输出错误",
			})
			return
		}
		testCaseBasic := &Models.TestCase{
			Identity:        Helper.GetUUID(),
			ProblemIdentity: identity,
			Input:           caseMap["input"],
			Output:          caseMap["output"],
		}
		testCaseBasics = append(testCaseBasics, testCaseBasic)
	}
	data.TestCases = testCaseBasics

	err := Models.CreateProblem(data).Error
	if err != nil {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "系统错误",
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
	fmt.Println("========== 进入 UpdateProblem 函数 ==========")

	identity := c.Query("identity")
	title := c.PostForm("title")
	content := c.PostForm("content")
	maxRuntime, _ := strconv.Atoi(c.PostForm("max_runtime"))
	maxMem, _ := strconv.Atoi(c.PostForm("max_mem"))
	categoryIds := c.PostFormArray("category_id")
	testCases := c.PostFormArray("test_cases")
	if identity == "" || title == "" || content == "" || len(categoryIds) == 0 || len(testCases) == 0 || maxRuntime == 0 || maxMem == 0 {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "参数不齐全",
		})
		return
	}

	err := Models.DB.Transaction(func(tx *gorm.DB) error {
		//问题基础信息的保存problem_basic

		problemBasic := &Models.ProblemBasic{
			Identity:   identity,
			Title:      title,
			Content:    content,
			MaxMem:     maxMem,
			MaxRuntime: maxRuntime,
		}
		err := tx.Model(new(Models.ProblemBasic)).
			Where("identity = ?", identity).
			Updates(problemBasic).Error
		if err != nil {
			return err
		}
		//查询问题详情
		err = tx.Where("identity = ?", identity).First(&problemBasic).Error
		if err != nil {
			return err
		}
		//关联问题分类的更新
		err = tx.Where("problem_id = ?", problemBasic.ID).Delete(&Models.ProblemCategory{}).Error
		if err != nil {
			return err
		}
		categoryBasics := make([]*Models.ProblemCategory, 0)
		for _, id := range categoryIds {
			atoi, _ := strconv.Atoi(id)
			categoryBasics = append(categoryBasics, &Models.ProblemCategory{
				CategoryId: uint(atoi),
				ProblemId:  problemBasic.ID,
			})
		}
		err = tx.Create(&categoryBasics).Error
		if err != nil {
			return err
		}
		//关联测试用例的更新
		err = tx.Where("problem_identity = ?", identity).Delete(&Models.TestCase{}).Error
		if err != nil {
			return err
		}
		testCaseBasics := make([]*Models.TestCase, 0)

		for _, testCase := range testCases {
			caseMap := make(map[string]string)
			err := json.Unmarshal([]byte(testCase), &caseMap)

			if err != nil {
				c.JSON(200, gin.H{
					"code": -1,
					"msg":  "参数错误",
				})
				return err
			}
			if _, ok := caseMap["input"]; !ok {
				c.JSON(200, gin.H{
					"code": -1,
					"msg":  "输入错误",
				})
				return err
			}
			if _, ok := caseMap["output"]; !ok {
				c.JSON(200, gin.H{
					"code": -1,
					"msg":  "输出错误",
				})
				return err
			}
			testCaseBasic := &Models.TestCase{
				Identity:        Helper.GetUUID(),
				ProblemIdentity: identity,
				Input:           caseMap["input"],
				Output:          caseMap["output"],
			}
			testCaseBasics = append(testCaseBasics, testCaseBasic)
		}
		err = tx.Create(&testCaseBasics).Error

		return nil
	})
	if err != nil {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "Problemupdate error",
		})
		return
	}
	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "修改成功",
	})
}
