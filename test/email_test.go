package test

import (
	"fmt"
	"net/smtp"
	"testing"

	"github.com/jordan-wright/email"
)

func TestEmail(t *testing.T) {
	e := email.NewEmail()
	// 1. 设置发件人、收件人、抄送/密送
	e.From = "Get <1151639091@qq.com>"
	e.To = []string{"2082209532@qq.com"}
	//e.Bcc = []string{"test_bcc@example.com"} // 密送
	//e.Cc = []string{"test_cc@example.com"}   // 抄送

	// 2. 设置邮件主题和内容
	e.Subject = "验证测试"
	e.Text = []byte("纯文本内容")               // 纯文本格式（备用）
	e.HTML = []byte("你的验证码是<b>123456</b>") // HTML 格式（优先）

	// 3. 连接 SMTP 服务器并发送邮件
	err := e.Send(
		"smtp.qq.com:587", // SMTP 服务器地址和端口
		smtp.PlainAuth(
			"",
			"1151639091@qq.com", // 发件邮箱账号
			"ztkuuqzjsltljdfe",  // 发件邮箱密码/授权码
			"smtp.qq.com",       // 服务器域名
		),
	)
	if err != nil {
		fmt.Println("发送邮件失败：", err)
	} else {
		fmt.Println("发送邮件成功！")
	}
}
