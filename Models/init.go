package Models

import (
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB = Init()

func Init() *gorm.DB {
	dsn := "root:123456@tcp(127.0.0.1:3306)/onlinepractice?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
	}
	return db
}
