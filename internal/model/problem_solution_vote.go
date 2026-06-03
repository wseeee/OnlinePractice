package model

import (
	"gorm.io/gorm"
	uuid "github.com/satori/go.uuid"
)

type ProblemSolutionVote struct {
	gorm.Model
	Identity        string `gorm:"column:identity;type:varchar(36);uniqueIndex" json:"identity"`
	ProblemIdentity string `gorm:"column:problem_identity;type:varchar(36);index" json:"problem_identity"`
	UserIdentity    string `gorm:"column:user_identity;type:varchar(36);index" json:"user_identity"`
	Vote            int    `gorm:"column:vote;type:tinyint" json:"vote"` // 1=赞, -1=踩
}

func (ProblemSolutionVote) TableName() string {
	return "problem_solution_vote"
}

// UpsertVote 投票（幂等：同一用户对同一题解只保留最后一次投票）
func UpsertVote(problemIdentity, userIdentity string, vote int) error {
	var existing ProblemSolutionVote
	err := DB.Where("problem_identity = ? AND user_identity = ?", problemIdentity, userIdentity).First(&existing).Error
	if err == nil {
		if existing.Vote == vote {
			// 重复投票视为取消
			return DB.Delete(&existing).Error
		}
		return DB.Model(&existing).Update("vote", vote).Error
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}
	v := &ProblemSolutionVote{
		Identity:        uuid.NewV4().String(),
		ProblemIdentity: problemIdentity,
		UserIdentity:    userIdentity,
		Vote:            vote,
	}
	return DB.Create(v).Error
}

func GetVoteCounts(problemIdentity string) (upvotes, downvotes int64) {
	DB.Model(new(ProblemSolutionVote)).Where("problem_identity = ? AND vote = 1", problemIdentity).Count(&upvotes)
	DB.Model(new(ProblemSolutionVote)).Where("problem_identity = ? AND vote = -1", problemIdentity).Count(&downvotes)
	return
}

func GetUserVote(problemIdentity, userIdentity string) int {
	var v ProblemSolutionVote
	err := DB.Where("problem_identity = ? AND user_identity = ?", problemIdentity, userIdentity).First(&v).Error
	if err != nil {
		return 0
	}
	return v.Vote
}
