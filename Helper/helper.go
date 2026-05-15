package Helper

import (
	"OnlinePrictice/define"
	"crypto/md5"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"net/smtp"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jordan-wright/email"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

// Md5 保留向后兼容（用于旧密码校验），新密码请使用 HashPassword
func Md5(str string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(str)))
}

// HashPassword 使用 bcrypt 哈希密码
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword 验证 bcrypt 密码
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

type UserJwt struct {
	Identity string `json:"identity"`
	Name     string `json:"name"`
	IsAdmin  int    `json:"is_admin"`
	jwt.RegisteredClaims
}

var myKey = func() []byte {
	key := os.Getenv("JWT_SECRET")
	if key == "" {
		key = "Online_Practice"
	}
	return []byte(key)
}()

func GenerateToken(identity, name string, isAdmin int) (string, error) {
	u := UserJwt{
		Identity: identity,
		Name:     name,
		IsAdmin:  isAdmin,
	}
	tokenstring := jwt.NewWithClaims(jwt.SigningMethodHS256, u)
	token, err := tokenstring.SignedString(myKey)
	if err != nil {
		log.Printf("Failed to generate token: %v", err)
		return "", err
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
		return nil, err
	}
	if !claims.Valid {
		return nil, fmt.Errorf("token invalid")
	}
	return user, nil
}

func SendEmail(useremail string, code string) error {
	smtpUser := os.Getenv("SMTP_USER")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpFrom := os.Getenv("SMTP_FROM")

	if smtpUser == "" {
		smtpUser = "1151639091@qq.com"
	}
	if smtpPassword == "" {
		return fmt.Errorf("SMTP_PASSWORD environment variable not set")
	}
	if smtpHost == "" {
		smtpHost = "smtp.qq.com"
	}
	if smtpPort == "" {
		smtpPort = "587"
	}
	if smtpFrom == "" {
		smtpFrom = smtpUser
	}

	e := email.NewEmail()
	e.From = "验证信息<" + smtpFrom + ">"
	e.To = []string{useremail}
	e.Subject = "验证码"
	e.Text = []byte("纯文本内容")
	e.HTML = []byte("验证码：<b>" + code + "</b>")

	return e.Send(
		smtpHost+":"+smtpPort,
		smtp.PlainAuth("", smtpUser, smtpPassword, smtpHost),
	)
}

func GetUUID() string {
	return uuid.NewV4().String()
}

// GenerateCode 随机生成6位验证码
func GenerateCode() string {
	max := big.NewInt(900000)
	n, _ := rand.Int(rand.Reader, max)
	code := 100000 + n.Int64()
	return fmt.Sprintf("%d", code)
}

// SaveCode 代码保存方法
func SaveCode(code []byte) (string, error) {
	dirname := "code/" + GetUUID()
	path := dirname + "/main.go"
	err := os.Mkdir(dirname, 0777)
	if err != nil {
		return "", err
	}
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	_, err = f.Write(code)
	if err != nil {
		return "", err
	}
	return path, nil
}

// 黑名单包 — 这些包绝不能出现在用户代码中
var dangerousPackages = map[string]struct{}{
	"os":          {},
	"os/exec":     {},
	"os/signal":   {},
	"net":         {},
	"net/http":    {},
	"net/smtp":    {},
	"net/url":     {},
	"syscall":     {},
	"unsafe":      {},
	"plugin":      {},
	"reflect":     {},
	"io":          {},
	"io/ioutil":   {},
	"path":        {},
	"path/filepath": {},
	"runtime/debug": {},
	"database/sql": {},
}

// CheckGoCodeValid 检查golang代码的合法性
func CheckGoCodeValid(path string) (bool, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	code := string(b)

	// 提取所有 import 路径
	imports := extractImports(code)
	for _, imp := range imports {
		// 检查是否在白名单
		if _, ok := define.ValidGolangPackageMap[imp]; ok {
			continue
		}
		// 检查是否在黑名单
		if _, ok := dangerousPackages[imp]; ok {
			return false, nil
		}
		// 不在白名单也不在黑名单 — 拒绝（默认拒绝策略）
		return false, nil
	}
	return true, nil
}

// extractImports 从 Go 源码中提取所有 import 路径
func extractImports(code string) []string {
	var imports []string
	lines := strings.Split(code, "\n")
	inBlock := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// 跳过注释行
		if strings.HasPrefix(trimmed, "//") {
			continue
		}

		if trimmed == "import (" {
			inBlock = true
			continue
		}
		if inBlock && trimmed == ")" {
			inBlock = false
			continue
		}

		if inBlock {
			imp := extractImportPath(trimmed)
			if imp != "" {
				imports = append(imports, imp)
			}
		} else if strings.HasPrefix(trimmed, "import ") {
			rest := strings.TrimSpace(trimmed[7:])
			imp := extractImportPath(rest)
			if imp != "" {
				imports = append(imports, imp)
			}
		}
	}
	return imports
}

// extractImportPath 从一行中提取 import 路径（处理别名和 _ 导入）
func extractImportPath(s string) string {
	// 去掉别名: "alias" 或 _ 或 .
	parts := strings.Fields(s)
	for _, p := range parts {
		p = strings.Trim(p, `"`)
		if p == "_" || p == "." || strings.HasSuffix(p, `"`) {
			continue
		}
		if strings.Contains(p, `"`) || strings.Contains(p, "/") || strings.Contains(p, ".") {
			// 这可能是包路径
			cleaned := strings.Trim(p, `"`)
			if cleaned != "" && !strings.HasPrefix(cleaned, "//") {
				return cleaned
			}
		}
	}
	// 最后一个字段通常是路径
	if len(parts) > 0 {
		last := strings.Trim(parts[len(parts)-1], `"`)
		if last != "" && last != "_" && last != "." && !strings.HasPrefix(last, "//") {
			return last
		}
	}
	return ""
}
