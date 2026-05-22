package Models

import "gorm.io/gorm"

type ContestProblem struct {
	gorm.Model
	ContestIdentity string        `gorm:"column:contest_identity;type:varchar(36);" json:"contest_identity"`
	ProblemIdentity string        `gorm:"column:problem_identity;type:varchar(36);" json:"problem_identity"`
	ProblemBasic    *ProblemBasic `gorm:"foreignKey:problem_identity;references:identity" json:"problem_basic"`
	Score           int           `gorm:"column:score;type:int(11);default:100;" json:"score"`
	Sort            int           `gorm:"column:sort;type:int(11);" json:"sort"`
}

func (table *ContestProblem) TableName() string {
	return "contest_problem"
}
