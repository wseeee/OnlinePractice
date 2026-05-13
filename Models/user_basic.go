package Models

import "gorm.io/gorm"

type UserBasic struct {
	gorm.Model
	Identity  string `gorm:"column:identity;type:varchar(36);" json:"identity"`
	Name      string `gorm:"column:name;type:varchar(100);" json:"name"`
	Password  string `gorm:"column:password;type:varchar(32);" json:"password"`
	Phone     string `gorm:"column:phone;type:char(11);" json:"phone"`
	Mail      string `gorm:"column:mail;type:varchar(100);" json:"mail"`
	PassNum   int    `gorm:"column:pass_num;type:int(11);" json:"pass_num"`
	SubmitNum int    `gorm:"column:submit_num;type:int(11);" json:"submit_num"`
	IsAdmin   int    `gorm:"column:is_admin;type:tinyint(1);" json:"is_admin"`
}

func (table *UserBasic) TableName() string {
	return "user_basic"
}

func GetUserList(identity string) *gorm.DB {
	return DB.Model(&UserBasic{}).
		Where("identity = ?", identity)
}

func UserLogin(u *UserBasic) *gorm.DB {
	return DB.Model(&UserBasic{}).
		Where("name = ? AND password = ?", u.Name, u.Password)
}

func InsertUser(u *UserBasic) *gorm.DB {
	return DB.Model(&UserBasic{}).Create(u)
}

func EmailExist(mail string) *gorm.DB {
	return DB.Model(&UserBasic{}).Where("mail = ?", mail)
}

func GetRankList() *gorm.DB {
	return DB.Model(&UserBasic{}).Order("finish_problem_num DESC,submit_num ASC")
}
