package router

import (
	_ "OnlinePrictice/docs"
	"OnlinePrictice/service"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Router() *gin.Engine {
	r := gin.Default()

	//配置swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/ping", service.Ping)
	//问题
	r.GET("/problem-list", service.GetProblemList)
	r.GET("/problem-detail", service.GetProblemDetail)

	//用户
	r.GET("/user-basic", service.GetUserBasic)
	r.POST("/login", service.Login)
	r.POST("/sendcode", service.SendCode)
	r.POST("/register", service.Register)
	//提交记录
	r.GET("/submit-list", service.GetSubmitList)
	return r
}
