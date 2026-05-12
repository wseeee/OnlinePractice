package test

import (
	"OnlinePrictice/Models"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()
var rdb = redis.NewClient(&redis.Options{
	Addr:     "192.168.64.128:6379", // 虚拟机IP + 端口
	Password: "123456",
	DB:       0, // 默认库
})

func TestRedisSET(t *testing.T) {
	rdb.Set(ctx, "name", "zhangsan", time.Second*20)

}

func TestRedisGET(t *testing.T) {
	val, err := rdb.Get(ctx, "2082209532@qq.com").Result()
	if err != nil {
		t.Log(err)
	}
	fmt.Println(val)
}

func TestRedisMOdel(t *testing.T) {
	v, err := Models.RDB.Get(ctx, "name").Result()
	if err != nil {
		t.Log(err)
	}
	fmt.Println(v)
}
