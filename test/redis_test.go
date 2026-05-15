package test

import (
	"OnlinePrictice/Models"
	"OnlinePrictice/define"
	"fmt"
	"testing"
	"time"
)

func TestRedisSET(t *testing.T) {
	err := Models.RDB.Set(define.CTX, "name", "zhangsan", time.Second*20).Err()
	if err != nil {
		t.Fatal(err)
	}
}

func TestRedisGET(t *testing.T) {
	val, err := Models.RDB.Get(define.CTX, "name").Result()
	if err != nil {
		t.Log(err)
	}
	fmt.Println(val)
}
