package Models

import (
	"gorm.io/gorm"
)

// Submit 提交记录结构体
type SubmitBasic struct {
	gorm.Model
	Identity        string        `gorm:"column:identity;type:varchar(36);" json:"identity"`
	ProblemIdentity string        `gorm:"column:problem_identity;type:varchar(36);" json:"problem_identity"`
	ProblemBasic    *ProblemBasic `gorm:"foreignKey:problem_identity;references:identity" json:"problem_basic"`
	UserIdentity    string        `gorm:"column:user_identity;type:varchar(36);" json:"user_identity"`
	UserBasic       *UserBasic    `gorm:"foreignKey:user_identity;references:identity" json:"user_basic"`
	Path            string        `gorm:"column:path;type:varchar(255);" json:"path"`
	Status          int           `gorm:"column:status;type:tinyint(1);" json:"status"`
}

func (table *SubmitBasic) TableName() string {
	return "submit_basic"
}

func GetSubmitList(problemidentity, useridentity string, status int) *gorm.DB {
	tx := DB.Model(&SubmitBasic{}).
		Preload("ProblemBasic", func(db *gorm.DB) *gorm.DB {
			return db.Omit("content")
		}).
		Preload("UserBasic", func(db *gorm.DB) *gorm.DB {
			return db.Omit("password")
		})
	if problemidentity != "" {
		tx = tx.Where("problem_identity = ?", problemidentity)
	}
	if useridentity != "" {
		tx = tx.Where("user_identity = ?", useridentity)
	}
	if status != 0 {
		tx = tx.Where("status = ?", status)
	}
	return tx
}
