package Models

import "gorm.io/gorm"

type ContestUser struct {
	gorm.Model
	ContestIdentity string `gorm:"column:contest_identity;type:varchar(36);" json:"contest_identity"`
	UserIdentity    string `gorm:"column:user_identity;type:varchar(36);" json:"user_identity"`
}

func (table *ContestUser) TableName() string {
	return "contest_user"
}

func IsContestUser(contestIdentity, userIdentity string) bool {
	var count int64
	DB.Model(new(ContestUser)).
		Where("contest_identity = ? AND user_identity = ?", contestIdentity, userIdentity).
		Count(&count)
	return count > 0
}

func GetContestUserCount(contestIdentity string) int64 {
	var count int64
	DB.Model(new(ContestUser)).
		Where("contest_identity = ?", contestIdentity).
		Count(&count)
	return count
}
