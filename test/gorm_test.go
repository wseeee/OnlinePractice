package test

import (
	"OnlinePrictice/Models"
	"fmt"
	"testing"
)

func TestGormTest(t *testing.T) {
	data := make([]*Models.ProblemBasic, 0)
	err := Models.DB.Find(&data).Error
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range data {
		fmt.Printf("%+v\n", v)
	}
}
