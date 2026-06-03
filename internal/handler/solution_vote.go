package handler

import (
	"OnlinePrictice/internal/helper"
	"OnlinePrictice/internal/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

// VoteSolution 对题解投票（用户 JWT）
func VoteSolution(c *gin.Context) {
	var req struct {
		ProblemIdentity string `json:"problem_identity"`
		Vote            int    `json:"vote"` // 1=赞, -1=踩
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.ProblemIdentity == "" || (req.Vote != 1 && req.Vote != -1) {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	u, _ := c.Get("user_claims")
	uid := u.(*helper.UserJwt).Identity

	if err := model.UpsertVote(req.ProblemIdentity, uid, req.Vote); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "投票失败"})
		return
	}

	// 返回最新计数
	up, down := model.GetVoteCounts(req.ProblemIdentity)
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{"upvotes": up, "downvotes": down},
	})
}

// GetSolutionVotes 获取题解投票数（公共）
func GetSolutionVotes(c *gin.Context) {
	problemIdentity := c.Query("problem_identity")
	if problemIdentity == "" {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "缺少 problem_identity"})
		return
	}
	up, down := model.GetVoteCounts(problemIdentity)
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{"upvotes": up, "downvotes": down},
	})
}

// GetMyVote 获取当前用户对题解的投票状态
func GetMyVote(c *gin.Context) {
	problemIdentity := c.Query("problem_identity")
	if problemIdentity == "" {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "缺少 problem_identity"})
		return
	}
	u, _ := c.Get("user_claims")
	uid := u.(*helper.UserJwt).Identity
	vote := model.GetUserVote(problemIdentity, uid)
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{"vote": vote}})
}
