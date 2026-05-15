package router

import (
	_ "OnlinePrictice/docs"
	"OnlinePrictice/middleware"
	"OnlinePrictice/service"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Router() *gin.Engine {
	r := gin.Default()

	//公有方法
	//配置swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	//问题
	r.GET("/problem-list", service.GetProblemList)
	r.GET("/problem-detail", service.GetProblemDetail)
	//用户
	r.GET("/user-basic", service.GetUserBasic)
	r.POST("/login", service.Login)
	r.POST("/sendcode", service.SendCode)
	r.POST("/register", service.Register)
	//排名
	r.GET("/rank-list", service.GetRankList)
	//提交记录
	r.GET("/submit-list", service.GetSubmitList)

	//用户私有方法
	//代码的提交判断
	user := r.Group("/user", middleware.AuthUserCheck())
	{
		user.POST("/code-submit", service.CodeSubmit)
	}

	//管理员私有方法
	//r.Group("/admin", middleware.AuthAdiminCheck())
	admin := r.Group("/admin", middleware.AuthAdiminCheck())
	{
		admin.POST("/problem-create", service.CreateProblem)
		admin.GET("/category-list", service.GetCategoryList)
		admin.POST("/category-create", service.CreateCategory)
		admin.DELETE("/category-delete", service.DeleteCategory)
		admin.PUT("/category-update", service.UpdateCategory)
		admin.PUT("/problem-update", service.UpdateProblem)
	}

	return r
}
