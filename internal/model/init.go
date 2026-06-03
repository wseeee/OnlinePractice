package model

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	DB  *gorm.DB
	RDB *redis.Client
)

func init() {
	loadEnv()
	DB = Init()
	RDB = InitRedisDB()
	InitMinIO()
	InitRabbitMQ()
}

func loadEnv() {
	// 向上查找 .env 文件（支持从子目录运行 go test）
	dir, _ := os.Getwd()
	for {
		envPath := filepath.Join(dir, ".env")
		if _, err := os.Stat(envPath); err == nil {
			_ = godotenv.Load(envPath)
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
}

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
		DB:       1,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Printf("Warning: Redis connection failed: %v (cache features degraded)", err)
	} else {
		log.Println("Redis connected successfully")
	}
	return rdb
}
