package service

import (
	"OnlinePrictice/Helper"
	"OnlinePrictice/Models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetUserBasic
// @Tags 公共方法
// @Summary 用户详情
// @Param identity query string false "useridentity "
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /user-basic [get]
func GetUserBasic(c *gin.Context) {
	identity := c.Query("identity")
	if identity == "" {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "参数不完整",
		})
		return
	}

	var data *Models.UserBasic
	err := Models.GetUserList(identity).Omit("Password").Find(&data).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(200, gin.H{
				"code": -1,
				"msg":  "用户不存在",
			})
			return
		}
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "用户故障",
		})
		return
	}
	c.JSON(200, gin.H{
		"code": 200,
		"data": data,
	})
}

// Login
// @Tags 公共方法
// @Summary 用户登录
// @Param username formData string false "用户名"
// @Param password formData string false "密码"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /login [post]
func Login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	if username == "" || password == "" {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "必填信息为空",
		})
		return
	}
	password = Helper.Md5(password)
	data := new(Models.UserBasic)
	data.Name = username
	data.Password = password
	
	err := Models.UserLogin(data).First(&data).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(200, gin.H{
				"code": -1,
				"msg":  "用户名或者密码错误",
			})
			return
		}
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "用户故障",
		})
		return
	}

	token, err := Helper.GenerateToken(data.Identity, data.Name)
	if err != nil {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "生成token出错",
		})
		return
	}
	c.JSON(200, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"token": token,
		},
	})
}
