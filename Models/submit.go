package Models

import "gorm.io/gorm"

// Submit 提交记录结构体
type Submit struct {
	gorm.Model
	Identity        string `gorm:"column:identity;type:varchar(36);" json:"identity"`
	ProblemIdentity string `gorm:"column:problem_identity;type:varchar(36);" json:"problem_identity"`
	UserIdentity    string `gorm:"column:user_identity;type:varchar(36);" json:"user_identity"`
	Path            string `gorm:"column:path;type:varchar(255);" json:"path"`
	Status          int8   `gorm:"column:status;type:tinyint(1);" json:"status"`
}

func (table *Submit) TableName() string {
	return "submit"
}
