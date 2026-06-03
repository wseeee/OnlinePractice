package router

import (
	_ "OnlinePrictice/docs"
	"OnlinePrictice/internal/handler"
	"OnlinePrictice/internal/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Router() *gin.Engine {
	r := gin.Default()

	// CORS 中间件
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	//公有方法
	//配置swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	//问题
	r.GET("/problem-list", handler.GetProblemList)
	r.GET("/problem-detail", handler.GetProblemDetail)
	//用户
	r.GET("/user-basic", handler.GetUserBasic)
	r.POST("/login", handler.Login)
	r.POST("/sendcode", handler.SendCode)
	r.POST("/register", handler.Register)
	//排名
	r.GET("/rank-list", handler.GetRankList)
	//提交记录
	r.GET("/submit-list", handler.GetSubmitList)
	//WebSocket
	r.GET("/ws", handler.WsHandler)
	//比赛
	r.GET("/contest-list", handler.GetContestListPublic)
	r.GET("/contest-detail", handler.GetContestDetailPublic)
	r.GET("/contest-rank", handler.GetContestRank)
	//题解
	r.GET("/problem-solution", handler.GetProblemSolution)
	//评论
	r.GET("/problem-comments", handler.GetProblemComments)
	//代码分享
	r.GET("/problem-code-shares", handler.GetProblemCodeShares)
	r.GET("/code-share-detail", handler.GetCodeShareDetail)
	//头像
	r.GET("/avatar-url", handler.GetAvatarURL)
	//签到排行榜
	r.GET("/check-in-leaderboard", handler.GetCheckInLeaderboard)
	//支持的语言
	r.GET("/languages", handler.GetLanguages)
	//标签
	r.GET("/tags", handler.GetTagList)
	r.GET("/problem-tags", handler.GetProblemTagList)
	//题解投票
	r.GET("/solution-votes", handler.GetSolutionVotes)

	//用户私有方法
	//代码的提交判断
	user := r.Group("/user", middleware.AuthUserCheck())
	{
		user.POST("/code-submit", handler.CodeSubmit)
		user.POST("/contest-join", handler.ContestJoin)
		user.POST("/contest-submit", handler.ContestSubmit)
		user.GET("/contest-submits", handler.GetUserContestSubmits)
		user.GET("/submit-result", handler.GetSubmitResult)
		//评论
		user.POST("/problem-comment", handler.PostProblemComment)
		user.DELETE("/problem-comment", handler.DeleteProblemComment)
		//代码分享
		user.POST("/code-share", handler.PostCodeShare)
		user.DELETE("/code-share", handler.DeleteCodeShare)
		//头像
		user.POST("/avatar-upload", handler.PostAvatarUpload)
		user.GET("/avatar-url", handler.GetMyAvatarURL)
		//签到
		user.POST("/check-in", handler.PostCheckIn)
		user.GET("/check-in-status", handler.GetCheckInStatus)
		user.GET("/check-in-calendar", handler.GetCheckInCalendar)
		//通知
		user.GET("/notifications", handler.GetNotifications)
		user.PUT("/notification-read", handler.MarkNotificationRead)
		user.PUT("/notifications-read-all", handler.MarkAllNotificationsRead)
		user.GET("/notification-unread-count", handler.GetUnreadNotificationCount)
		//题解投票
		user.POST("/solution-vote", handler.VoteSolution)
		user.GET("/my-vote", handler.GetMyVote)
		//收藏
		user.POST("/problem-favorite", handler.AddFavorite)
		user.DELETE("/problem-favorite", handler.RemoveFavorite)
		user.GET("/problem-favorites", handler.GetFavorites)
		user.GET("/problem-favorite-check", handler.CheckFavorite)
		//做题笔记
		user.POST("/problem-note", handler.SaveProblemNote)
		user.GET("/problem-note", handler.GetProblemNote)
		user.DELETE("/problem-note", handler.DeleteProblemNote)
	}

	//管理员私有方法
	admin := r.Group("/admin", middleware.AuthAdminCheck())
	{
		admin.POST("/problem-create", handler.CreateProblem)
		admin.GET("/category-list", handler.GetCategoryList)
		admin.POST("/category-create", handler.CreateCategory)
		admin.DELETE("/category-delete", handler.DeleteCategory)
		admin.PUT("/category-update", handler.UpdateCategory)
		admin.PUT("/problem-update", handler.UpdateProblem)
		admin.POST("/contest-create", handler.CreateContest)
		admin.PUT("/contest-update", handler.UpdateContest)
		admin.GET("/contest-list", handler.GetContestList)
		admin.DELETE("/contest-delete", handler.DeleteContest)
		//题解
		admin.PUT("/problem-solution", handler.UpsertProblemSolution)
		admin.DELETE("/problem-solution", handler.DeleteProblemSolution)
		//评论管理
		admin.DELETE("/problem-comment", handler.AdminHideProblemComment)
		//标签管理
		admin.POST("/tag-create", handler.CreateTag)
		admin.PUT("/tag-update", handler.UpdateTag)
		admin.DELETE("/tag-delete", handler.DeleteTag)
		admin.POST("/problem-tags-set", handler.SetProblemTags)
	}

	return r
}
