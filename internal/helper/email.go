package helper

import (
	"fmt"
	"net/smtp"
	"os"
	"regexp"

	"github.com/jordan-wright/email"
)

// EmailConfig 邮件服务配置
type EmailConfig struct {
	Backend      string // smtp
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string
	SenderName   string
}

// EmailProvider 预设邮件服务商
type EmailProvider struct {
	Name     string
	SMTPHost string
	SMTPPort string
	Desc     string
}

// ProviderPresets 预设邮箱服务商，通过 EMAIL_PROVIDER 环境变量切换
var ProviderPresets = map[string]EmailProvider{
	"qq": {
		Name:     "QQ邮箱",
		SMTPHost: "smtp.qq.com",
		SMTPPort: "587",
		Desc:     "QQ邮箱 SMTP，需在设置→账户→开启POP3/SMTP服务，使用授权码而非QQ密码",
	},
	"163": {
		Name:     "网易163邮箱",
		SMTPHost: "smtp.163.com",
		SMTPPort: "465",
		Desc:     "163邮箱 SMTP，需在设置→POP3/SMTP/IMAP→开启SMTP服务，使用授权码",
	},
	"gmail": {
		Name:     "Google Gmail",
		SMTPHost: "smtp.gmail.com",
		SMTPPort: "587",
		Desc:     "Gmail SMTP，需开启两步验证后生成应用专用密码",
	},
	"outlook": {
		Name:     "Outlook / Exchange Online",
		SMTPHost: "smtp-mail.outlook.com",
		SMTPPort: "587",
		Desc:     "Microsoft 365 / Exchange Online SMTP，使用邮箱密码或应用密码",
	},
	"aliyun": {
		Name:     "阿里企业邮箱",
		SMTPHost: "smtp.mxhichina.com",
		SMTPPort: "465",
		Desc:     "阿里云企业邮箱 SMTP",
	},
}

// LoadEmailConfig 从环境变量加载邮件配置
func LoadEmailConfig() *EmailConfig {
	cfg := &EmailConfig{
		Backend:    envDefault("EMAIL_BACKEND", "smtp"),
		SenderName: envDefault("EMAIL_SENDER_NAME", "OnlineJudge"),
	}

	// 预设服务商快捷配置
	preset := os.Getenv("EMAIL_PROVIDER")
	if p, ok := ProviderPresets[preset]; ok {
		cfg.SMTPHost = envDefault("SMTP_HOST", p.SMTPHost)
		cfg.SMTPPort = envDefault("SMTP_PORT", p.SMTPPort)
	} else {
		cfg.SMTPHost = envDefault("SMTP_HOST", "smtp.qq.com")
		cfg.SMTPPort = envDefault("SMTP_PORT", "587")
	}

	cfg.SMTPUser = envDefault("SMTP_USER", "1151639091@qq.com")
	cfg.SMTPPassword = os.Getenv("SMTP_PASSWORD")
	cfg.SMTPFrom = envDefault("SMTP_FROM", cfg.SMTPUser)

	return cfg
}

// SendVerificationCode 发送验证码邮件
func (cfg *EmailConfig) SendVerificationCode(toEmail, code string) error {
	if cfg.Backend != "smtp" {
		return fmt.Errorf("不支持的后端: %s，当前仅支持 smtp", cfg.Backend)
	}
	if cfg.SMTPPassword == "" {
		return fmt.Errorf("SMTP_PASSWORD 未设置，请在 .env 中配置")
	}

	from := fmt.Sprintf("%s <%s>", cfg.SenderName, cfg.SMTPFrom)

	e := email.NewEmail()
	e.From = from
	e.To = []string{toEmail}
	e.Subject = fmt.Sprintf("【%s】邮箱验证码", cfg.SenderName)
	e.Text = []byte(fmt.Sprintf("您的验证码是：%s，5分钟内有效。请勿泄露给他人。", code))
	e.HTML = []byte(buildVerificationHTML(cfg.SenderName, code))

	addr := cfg.SMTPHost + ":" + cfg.SMTPPort
	auth := smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPHost)
	return e.Send(addr, auth)
}

func envDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// ValidateEmail 校验邮箱格式
func ValidateEmail(email string) bool {
	if len(email) > 254 {
		return false
	}
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

func buildVerificationHTML(appName, code string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="font-family:Arial,sans-serif;background:#f5f5f5;padding:20px;">
  <div style="max-width:480px;margin:0 auto;background:#fff;border-radius:8px;padding:32px;box-shadow:0 1px 4px rgba(0,0,0,.1);">
    <h2 style="color:#333;margin-top:0;">%s</h2>
    <p style="color:#666;line-height:1.6;">您正在注册或登录，请使用以下验证码完成验证：</p>
    <div style="background:#f0f7ff;border:1px solid #d0e3f7;border-radius:6px;padding:16px;text-align:center;margin:20px 0;">
      <span style="font-size:28px;font-weight:bold;letter-spacing:6px;color:#1890ff;">%s</span>
    </div>
    <p style="color:#999;font-size:13px;line-height:1.6;">验证码 5 分钟内有效，请勿泄露给他人。<br>如非本人操作，请忽略此邮件。</p>
  </div>
</body>
</html>`, appName, code)
}
