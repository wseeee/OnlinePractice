package model

import (
	"log"

	"gorm.io/gorm"
)

type CodeShare struct {
	gorm.Model
	Identity        string `gorm:"column:identity;type:varchar(36);uniqueIndex" json:"identity"`
	SubmitIdentity  string `gorm:"column:submit_identity;type:varchar(36);uniqueIndex" json:"submit_identity"`
	ProblemIdentity string `gorm:"column:problem_identity;type:varchar(36);index" json:"problem_identity"`
	UserIdentity    string `gorm:"column:user_identity;type:varchar(36);" json:"user_identity"`
	Title           string `gorm:"column:title;type:varchar(255);" json:"title"`
	ViewCount       int    `gorm:"column:view_count;type:int(11);default:0" json:"view_count"`
}

func (table *CodeShare) TableName() string {
	return "code_share"
}

func CreateShare(s *CodeShare) error {
	return DB.Create(s).Error
}

func DeleteShareByIdentity(identity string) error {
	return DB.Where("identity = ?", identity).Delete(&CodeShare{}).Error
}

func ListSharesByProblem(problemIdentity string, page, size int) ([]*CodeShare, int64, error) {
	var shares []*CodeShare
	var count int64
	q := DB.Model(new(CodeShare)).Where("problem_identity = ?", problemIdentity)
	if err := q.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	err := q.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&shares).Error
	return shares, count, err
}

func GetShareByIdentity(identity string) (*CodeShare, error) {
	var s CodeShare
	err := DB.Where("identity = ?", identity).First(&s).Error
	return &s, err
}

func ExistsShareBySubmit(submitIdentity string) bool {
	var count int64
	if err := DB.Model(new(CodeShare)).Where("submit_identity = ?", submitIdentity).Count(&count).Error; err != nil {
		log.Printf("ExistsShareBySubmit error: %v", err)
		return false
	}
	return count > 0
}
