package service

import (
	"OnlinePrictice/Helper"
	"OnlinePrictice/Models"
	"OnlinePrictice/define"
	"log"
	"strconv"
	"time"

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
			"msg":  "查询失败",
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
			"msg":  "必填信息不能为空",
		})
		return
	}

	// 先按用户名查询用户
	data := new(Models.UserBasic)
	err := Models.DB.Where("name = ?", username).First(data).Error
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
			"msg":  "查询失败",
		})
		return
	}

	// 验证密码（兼容 bcrypt 和旧的 MD5）
	if !Helper.CheckPassword(password, data.Password) {
		// 兼容旧密码: 尝试 MD5 比对
		if data.Password != Helper.Md5(password) {
			c.JSON(200, gin.H{
				"code": -1,
				"msg":  "用户名或者密码错误",
			})
			return
		}
		// 旧密码匹配，自动升级为 bcrypt
		newHash, _ := Helper.HashPassword(password)
		Models.DB.Model(data).Update("password", newHash)
	}

	token, err := Helper.GenerateToken(data.Identity, data.Name, data.IsAdmin)
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
			"token":   token,
			"isAdmin": data.IsAdmin,
		},
	})
}

// SendCode
// @Tags 公共方法
// @Summary 发送验证码
// @Param mail formData string false "mail"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /sendcode [post]
func SendCode(c *gin.Context) {
	mail := c.PostForm("mail")
	if mail == "" {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "邮箱不能为空",
		})
		return
	}
	code := Helper.GenerateCode()
	Models.RDB.Set(define.CTX, mail, code, time.Second*300)
	err := Helper.SendEmail(mail, code)
	if err != nil {
		log.Printf("发送邮件失败: %v", err)
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "发送邮件失败",
		})
		return
	}
	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "发送成功",
	})
}

// Register
// @Tags 公共方法
// @Summary 用户的注册
// @Param name formData string true "用户名"
// @Param password formData string true "密码"
// @Param phone formData string false "phone"
// @Param mail formData string true "mail"
// @Param code formData string true "code"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /register [post]
func Register(c *gin.Context) {
	name := c.PostForm("name")
	password := c.PostForm("password")
	mail := c.PostForm("mail")
	phone := c.PostForm("phone")
	code := c.PostForm("code")
	if name == "" || password == "" || mail == "" || phone == "" || code == "" {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "必填信息不能为空",
		})
		return
	}

	//验证码是否正确
	GetCode, err := Models.RDB.Get(define.CTX, mail).Result()
	if err != nil {
		log.Printf("获取验证码失败: %v", err)
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "验证码已过期",
		})
		return
	}
	if GetCode != code {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "验证码错误",
		})
		return
	}
	//判断邮箱是否存在
	var count int64
	err = Models.EmailExist(mail).Count(&count).Error
	if err != nil {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "服务器异常，请稍后再试",
		})
		return
	}
	if count > 0 {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "该邮箱已注册",
		})
		return
	}
	//数据的插入
	user := new(Models.UserBasic)
	user.Name = name
	hashedPwd, err := Helper.HashPassword(password)
	if err != nil {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "密码处理失败",
		})
		return
	}
	user.Password = hashedPwd
	user.Identity = Helper.GetUUID()
	user.Mail = mail
	user.Phone = phone

	err = Models.InsertUser(user).Error
	if err != nil {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "用户创建失败",
		})
		return
	}

	token, err := Helper.GenerateToken(user.Identity, user.Name, user.IsAdmin)
	if err != nil {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "生成token出错",
		})
		return
	}
	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "用户创建成功",
		"data": map[string]interface{}{
			"token": token,
		},
	})
}

// GetRankList
// @Tags 公共方法
// @Summary 排名
// @Param page query int false "请输入当前页，默认第一页"
// @Param size query int false "size"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /rank-list [get]
func GetRankList(c *gin.Context) {
	size, _ := strconv.Atoi(c.DefaultQuery("size", define.DefaultSize))
	page, err := strconv.Atoi(c.DefaultQuery("page", define.DefaultPage))
	if err != nil {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "参数错误",
		})
		return
	}
	page = (page - 1) * size
	var count int64
	list := make([]*Models.UserBasic, 0)
	// 先 Count 再分页查询
	err = Models.GetRankList().Count(&count).Offset(page).Limit(size).Find(&list).Error
	if err != nil {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "获取数据失败",
		})
		return
	}
	c.JSON(200, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"count": count,
			"list":  list,
		},
	})
}
