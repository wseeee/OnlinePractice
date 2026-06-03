package model

import (
	"time"

	"gorm.io/gorm"
)

type UserCheckIn struct {
	gorm.Model
	Identity     string `gorm:"column:identity;type:varchar(36);uniqueIndex" json:"identity"`
	UserIdentity string `gorm:"column:user_identity;type:varchar(36);index" json:"user_identity"`
	CheckDate    string `gorm:"column:check_date;type:date;" json:"check_date"`
	Streak       int    `gorm:"column:streak;type:int(11);" json:"streak"`
	Points       int    `gorm:"column:points;type:int(11);" json:"points"`
}

func (table *UserCheckIn) TableName() string {
	return "user_check_in"
}

func GetCheckInDates(userIdentity string, year, month int) ([]string, error) {
	var dates []string
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0)
	err := DB.Model(&UserCheckIn{}).
		Where("user_identity = ? AND check_date >= ? AND check_date < ?", userIdentity, start.Format("2006-01-02"), end.Format("2006-01-02")).
		Pluck("check_date", &dates).Error
	return dates, err
}

func GetCheckInLeaderboard(limit int) ([]UserBasic, error) {
	var users []UserBasic
	err := DB.Model(&UserBasic{}).
		Where("check_in_points > 0").
		Order("check_in_points DESC").
		Limit(limit).
		Find(&users).Error
	return users, err
}
