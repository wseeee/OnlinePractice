package model

import "gorm.io/gorm"

type ContestSubmit struct {
	gorm.Model
	Identity        string        `gorm:"column:identity;type:varchar(36);uniqueIndex" json:"identity"`
	ContestIdentity string        `gorm:"column:contest_identity;type:varchar(36);" json:"contest_identity"`
	ProblemIdentity string        `gorm:"column:problem_identity;type:varchar(36);" json:"problem_identity"`
	UserIdentity    string        `gorm:"column:user_identity;type:varchar(36);" json:"user_identity"`
	Path            string        `gorm:"column:path;type:varchar(255);" json:"path"`
	Score           int           `gorm:"column:score;type:int(11);" json:"score"`
	Status          int           `gorm:"column:status;type:tinyint(1);" json:"status"`
	Language        string        `gorm:"column:language;type:varchar(32);default:go" json:"language"`
	ProblemBasic    *ProblemBasic `gorm:"foreignKey:problem_identity;references:identity" json:"problem_basic"`
	UserBasic       *UserBasic    `gorm:"foreignKey:user_identity;references:identity" json:"user_basic"`
}

func (table *ContestSubmit) TableName() string {
	return "contest_submit"
}

func GetContestSubmitList(contestIdentity, problemIdentity, userIdentity string) *gorm.DB {
	tx := DB.Model(new(ContestSubmit)).
		Where("contest_identity = ?", contestIdentity).
		Preload("ProblemBasic", func(db *gorm.DB) *gorm.DB {
			return db.Omit("content")
		}).
		Preload("UserBasic", func(db *gorm.DB) *gorm.DB {
			return db.Omit("password")
		})
	if problemIdentity != "" {
		tx = tx.Where("problem_identity = ?", problemIdentity)
	}
	if userIdentity != "" {
		tx = tx.Where("user_identity = ?", userIdentity)
	}
	return tx.Order("created_at DESC")
}
