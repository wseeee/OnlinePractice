package Helper

import (
	"crypto/md5"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"net/smtp"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jordan-wright/email"
	uuid "github.com/satori/go.uuid"
)

func Md5(str string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(str)))
}

type UserJwt struct {
	Identity string `json:"identity"`
	Name     string `json:"name"`
	jwt.RegisteredClaims
}

var myKey = []byte("Online_Practice")

func GenerateToken(identity, name string) (string, error) {
	u := UserJwt{
		Identity: identity,
		Name:     name,
	}
	tokenstring := jwt.NewWithClaims(jwt.SigningMethodHS256, u)
	token, err := tokenstring.SignedString(myKey)
	if err != nil {
		log.Print("Failed to generate token: %v", err)
	}
	return token, nil
}

func AnalyseToken(token string) (*UserJwt, error) {
	user := new(UserJwt)
	claims, err := jwt.ParseWithClaims(token, user,
		func(token *jwt.Token) (interface{}, error) {
			return myKey, nil
		})
	if err != nil {
		log.Print("Failed to analyse token: %v", err)
	}
	if claims.Valid {
		return nil, fmt.Errorf("Analyse Token Error")
	}
	return user, nil
}

func SendEmail(useremail string, code string) error {
	e := email.NewEmail()
	// 1. 设置发件人、收件人、抄送/密送
	e.From = "验证信息<1151639091@qq.com>"
	e.To = []string{useremail}
	//e.Bcc = []string{"test_bcc@example.com"} // 密送
	//e.Cc = []string{"test_cc@example.com"}   // 抄送

	// 2. 设置邮件主题和内容
	e.Subject = "验证码"
	e.Text = []byte("纯文本内容")                 // 纯文本格式（备用）
	e.HTML = []byte("验证码：<b>" + code + "</b>") // HTML 格式（优先）

	// 3. 连接 SMTP 服务器并发送邮件
	return e.Send(
		"smtp.qq.com:587", // SMTP 服务器地址和端口
		smtp.PlainAuth(
			"",
			"1151639091@qq.com", // 发件邮箱账号
			"ztkuuqzjsltljdfe",  // 发件邮箱密码/授权码
			"smtp.qq.com",       // 服务器域名
		),
	)

}

func GetUUID() string {
	return uuid.NewV4().String()
}

// 随机生成code
func GenerateCode() string {
	max := big.NewInt(900000)
	n, _ := rand.Int(rand.Reader, max)
	code := 100000 + n.Int64()
	return fmt.Sprintf("%d", code)
}
