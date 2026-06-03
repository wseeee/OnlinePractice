package model

import (
	"gorm.io/gorm"
)

type ProblemSolution struct {
	gorm.Model
	Identity         string `gorm:"column:identity;type:varchar(36);uniqueIndex" json:"identity"`
	ProblemIdentity  string `gorm:"column:problem_identity;type:varchar(36);index" json:"problem_identity"`
	Title            string `gorm:"column:title;type:varchar(255);" json:"title"`
	Content          string `gorm:"column:content;type:text;" json:"content"`
	Language         string `gorm:"column:language;type:varchar(32);default:go" json:"language"`
	AuthorIdentity   string `gorm:"column:author_identity;type:varchar(36);" json:"author_identity"`
	IsPublished      int    `gorm:"column:is_published;type:tinyint;default:1" json:"is_published"`
}

func (table *ProblemSolution) TableName() string {
	return "problem_solution"
}

func GetPublishedSolution(problemIdentity string) (*ProblemSolution, error) {
	var s ProblemSolution
	err := DB.Where("problem_identity = ? AND is_published = 1", problemIdentity).First(&s).Error
	return &s, err
}

func UpsertSolution(s *ProblemSolution) error {
	var existing ProblemSolution
	err := DB.Where("problem_identity = ?", s.ProblemIdentity).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return DB.Create(s).Error
	}
	if err != nil {
		return err
	}
	return DB.Model(&existing).Updates(map[string]interface{}{
		"title":        s.Title,
		"content":      s.Content,
		"language":     s.Language,
		"author_identity": s.AuthorIdentity,
		"is_published": s.IsPublished,
	}).Error
}

func DeleteSolution(problemIdentity string) error {
	return DB.Where("problem_identity = ?", problemIdentity).Delete(&ProblemSolution{}).Error
}
