package Models

import (
	"log"

	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB = Init()
var RDB = InitRedis()

func Init() *gorm.DB {
	dsn := "root:123456@tcp(127.0.0.1:3306)/onlinepractice?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
	}
	return db
}

func InitRedis() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "192.168.64.128:6379", // 虚拟机IP + 端口
		Password: "123456",
		DB:       0,
	})
	return rdb
}
