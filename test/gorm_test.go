package test

import (
	"OnlinePrictice/Models"
	"fmt"
	"testing"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestGormTest(t *testing.T) {
	dsn := "root:123456@tcp(127.0.0.1:3306)/onlinepractice?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		t.Fatal(err)
	}

	data := make([]*Models.Problem, 0)
	err = db.Find(&data).Error
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range data {
		fmt.Printf("%+v\n", v)
	}
}
