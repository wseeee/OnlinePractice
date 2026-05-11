package Models

import "gorm.io/gorm"

type Problem struct {
	gorm.Model
	Identity   string `gorm:"column:identity;type:varchar(36);" json:"identity"`
	CategroyId string `gorm:"column:categroy_id;type:varchar(255);" json:"categroy_id"`
	Title      string `gorm:"column:title;type:varchar(255);" json:"title"`
	Content    string `gorm:"column:content;type:text;" json:"content"`
}

func (table *Problem) TableName() string {
	return "problem"
}
