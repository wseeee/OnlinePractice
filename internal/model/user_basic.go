package model

import (
	"database/sql"
	"gorm.io/gorm"
)

type UserBasic struct {
	gorm.Model
	Identity       string        `gorm:"column:identity;type:varchar(36);" json:"identity"`
	Name           string        `gorm:"column:name;type:varchar(100);" json:"name"`
	Password       string        `gorm:"column:password;type:varchar(255);" json:"-"`
	Phone          string        `gorm:"column:phone;type:char(11);" json:"phone"`
	Mail           string        `gorm:"column:mail;type:varchar(100);" json:"mail"`
	AvatarKey      string        `gorm:"column:avatar_key;type:varchar(255);" json:"avatar_key"`
	PassNum        int           `gorm:"column:pass_num;type:int(11);" json:"pass_num"`
	SubmitNum      int           `gorm:"column:submit_num;type:int(11);" json:"submit_num"`
	CheckInTotal   int           `gorm:"column:check_in_total;type:int(11);" json:"check_in_total"`
	CheckInStreak  int           `gorm:"column:check_in_streak;type:int(11);" json:"check_in_streak"`
	CheckInPoints  int           `gorm:"column:check_in_points;type:int(11);" json:"check_in_points"`
	LastCheckInDate sql.NullTime `gorm:"column:last_check_in_date;type:date;" json:"last_check_in_date"`
	IsAdmin        int          `gorm:"column:is_admin;type:tinyint(1);" json:"is_admin"`
}

func (table *UserBasic) TableName() string {
	return "user_basic"
}

func GetUserList(identity string) *gorm.DB {
	return DB.Model(&UserBasic{}).
		Where("identity = ?", identity)
}

func InsertUser(u *UserBasic) *gorm.DB {
	return DB.Model(&UserBasic{}).Create(u)
}

func EmailExist(mail string) *gorm.DB {
	return DB.Model(&UserBasic{}).Where("mail = ?", mail)
}

func GetRankList() *gorm.DB {
	return DB.Model(&UserBasic{}).Omit("password").Order("pass_num DESC,submit_num ASC")
}
