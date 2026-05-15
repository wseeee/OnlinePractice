package test

import (
	"OnlinePrictice/Helper"
	"fmt"
	"testing"
)

func TestEmail(t *testing.T) {
	err := Helper.SendEmail("2082209532@qq.com", "123456")
	if err != nil {
		fmt.Println("发送邮件失败：", err)
	} else {
		fmt.Println("发送邮件成功！")
	}
}
