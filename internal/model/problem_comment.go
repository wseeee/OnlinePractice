package model

import (
	"gorm.io/gorm"
)

type ProblemComment struct {
	gorm.Model
	Identity        string `gorm:"column:identity;type:varchar(36);uniqueIndex" json:"identity"`
	ProblemIdentity string `gorm:"column:problem_identity;type:varchar(36);index" json:"problem_identity"`
	UserIdentity    string `gorm:"column:user_identity;type:varchar(36);" json:"user_identity"`
	ParentIdentity  string `gorm:"column:parent_identity;type:varchar(36);index;default:" json:"parent_identity"`
	Content         string `gorm:"column:content;type:text;" json:"content"`
	Status          int    `gorm:"column:status;type:tinyint;default:1" json:"status"`
	UserName        string `gorm:"-" json:"user_name"`
	Replies         []*ProblemComment `gorm:"-" json:"replies,omitempty"`
}

func (table *ProblemComment) TableName() string {
	return "problem_comment"
}

func CreateComment(c *ProblemComment) error {
	return DB.Create(c).Error
}

func ListTopLevelComments(problemIdentity string, page, size int) ([]*ProblemComment, int64, error) {
	var comments []*ProblemComment
	var count int64
	q := DB.Model(new(ProblemComment)).
		Where("problem_identity = ? AND parent_identity = '' AND status = 1", problemIdentity)
	if err := q.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	err := q.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&comments).Error
	return comments, count, err
}

func ListRepliesByParentIdentities(parentIdentities []string) ([]*ProblemComment, error) {
	if len(parentIdentities) == 0 {
		return nil, nil
	}
	var replies []*ProblemComment
	err := DB.Where("parent_identity IN ? AND status = 1", parentIdentities).
		Order("created_at ASC").Find(&replies).Error
	return replies, err
}

func GetCommentByIdentity(identity string) (*ProblemComment, error) {
	var c ProblemComment
	err := DB.Where("identity = ?", identity).First(&c).Error
	return &c, err
}

func SoftDeleteComment(identity string) error {
	return DB.Where("identity = ?", identity).Delete(&ProblemComment{}).Error
}

func AdminHideComment(identity string) error {
	return DB.Model(new(ProblemComment)).Where("identity = ?", identity).Update("status", 0).Error
}
