package Models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name     string `gorm:"column:name;type:varchar(100);" json:"name"`
	Password string `gorm:"column:password;type:varchar(32);" json:"password"`
	Phone    string `gorm:"column:phone;type:char(11);" json:"phone"`
	Mail     string `gorm:"column:mail;type:varchar(100);" json:"mail"`
}

func (table *User) TableName() string {
	return "user"
}
