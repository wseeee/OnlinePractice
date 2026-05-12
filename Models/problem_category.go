package Models

import "gorm.io/gorm"

type ProblemCategory struct {
	gorm.Model
	ProblemId     int            `gorm:"column:problem_id;" json:"problem_id"`
	CategoryId    int            `gorm:"column:category_id;" json:"category_id"`
	CategoryBasic *CategoryBasic `gorm:"foreignKey:id;references:category_id" json:"category_basic"`
}

func (table *ProblemCategory) TableName() string {
	return "problem_category"
}
