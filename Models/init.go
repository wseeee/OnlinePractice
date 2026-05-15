package Models

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB = Init()
var RDB = InitRedisDB()

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func Init() *gorm.DB {
	host := getEnv("DB_HOST", "127.0.0.1")
	port := getEnv("DB_PORT", "3306")
	user := getEnv("DB_USER", "root")
	password := getEnv("DB_PASSWORD", "123456")
	dbname := getEnv("DB_NAME", "onlinepractice")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, password, host, port, dbname)

	var db *gorm.DB
	var err error
	for i := 0; i < 30; i++ {
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err == nil {
			log.Println("Database connected successfully")
			return db
		}
		log.Printf("Database not ready (attempt %d/30): %v", i+1, err)
		time.Sleep(2 * time.Second)
	}
	log.Fatalf("Failed to connect to database after 30 attempts: %v", err)
	return nil
}

func InitRedisDB() *redis.Client {
	host := getEnv("REDIS_HOST", "127.0.0.1")
	port := getEnv("REDIS_PORT", "6379")
	password := getEnv("REDIS_PASSWORD", "123456")

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       0,
	})

	for i := 0; i < 10; i++ {
		if err := rdb.Ping(rdb.Context()).Err(); err == nil {
			log.Println("Redis connected successfully")
			return rdb
		}
		log.Printf("Redis not ready (attempt %d/10), retrying...", i+1)
		time.Sleep(2 * time.Second)
	}
	log.Println("Warning: Redis connection could not be established")
	return rdb
}
