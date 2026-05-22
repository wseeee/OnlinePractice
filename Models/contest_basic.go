package Models

import (
	"time"

	"gorm.io/gorm"
)

type ContestBasic struct {
	gorm.Model
	Identity        string            `gorm:"column:identity;type:varchar(36);" json:"identity"`
	Title           string            `gorm:"column:title;type:varchar(255);" json:"title"`
	Description     string            `gorm:"column:description;type:text;" json:"description"`
	StartAt         time.Time         `gorm:"column:start_at;" json:"start_at"`
	EndAt           time.Time         `gorm:"column:end_at;" json:"end_at"`
	ContestType     int               `gorm:"column:contest_type;type:tinyint(1);" json:"contest_type"`
	MaxParticipants int               `gorm:"column:max_participants;type:int(11);default:0;" json:"max_participants"`
	ContestProblems []*ContestProblem `gorm:"foreignKey:contest_identity;references:identity" json:"contest_problems"`
}

func (table *ContestBasic) TableName() string {
	return "contest_basic"
}

func GetContestList(keyword string) *gorm.DB {
	return DB.Model(new(ContestBasic)).
		Where("title like ?", "%"+keyword+"%").
		Order("start_at DESC")
}

func GetContestDetail(identity string) *gorm.DB {
	return DB.Model(new(ContestBasic)).
		Preload("ContestProblems").
		Preload("ContestProblems.ProblemBasic").
		Where("identity = ?", identity)
}

func CreateContest(c *ContestBasic) *gorm.DB {
	return DB.Model(new(ContestBasic)).Create(c)
}

func DeleteContest(identity string) *gorm.DB {
	return DB.Model(new(ContestBasic)).
		Where("identity = ?", identity).Delete(new(ContestBasic))
}

func UpdateContest(identity string, c *ContestBasic) *gorm.DB {
	return DB.Model(new(ContestBasic)).
		Where("identity = ?", identity).Updates(c)
}
