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
	r.GET("/problem-list", service.GetProblemList)
	return r
}
