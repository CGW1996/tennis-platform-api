package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"tennis-platform/backend/internal/config"

	"gopkg.in/gomail.v2"
)

// EmailService 郵件服務
type EmailService struct {
	config *config.Config
	dialer *gomail.Dialer
}

// NewEmailService 創建新的郵件服務
func NewEmailService(cfg *config.Config) *EmailService {
	// 這裡使用基本配置，實際部署時需要配置真實的 SMTP 服務器
	dialer := gomail.NewDialer("localhost", 587, "", "")

	return &EmailService{
		config: cfg,
		dialer: dialer,
	}
}

// SendVerificationEmail 發送驗證郵件
func (e *EmailService) SendVerificationEmail(email, token string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", "noreply@tennis-platform.com")
	m.SetHeader("To", email)
	m.SetHeader("Subject", "驗證您的電子郵件地址")

	verificationURL := fmt.Sprintf("http://localhost:3000/verify-email?token=%s", token)

	body := fmt.Sprintf(`
		<h2>歡迎加入網球平台！</h2>
		<p>請點擊下面的連結來驗證您的電子郵件地址：</p>
		<a href="%s">驗證電子郵件</a>
		<p>如果您沒有註冊帳號，請忽略此郵件。</p>
		<p>此連結將在24小時後過期。</p>
	`, verificationURL)

	m.SetBody("text/html", body)

	// 在開發環境中，我們只是記錄郵件內容而不實際發送
	if e.config.Env == "development" {
		fmt.Printf("Email would be sent to %s with verification URL: %s\n", email, verificationURL)
		return nil
	}

	return e.dialer.DialAndSend(m)
}

// SendPasswordResetEmail 發送密碼重設郵件
func (e *EmailService) SendPasswordResetEmail(email, token string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", "noreply@tennis-platform.com")
	m.SetHeader("To", email)
	m.SetHeader("Subject", "重設您的密碼")

	resetURL := fmt.Sprintf("http://localhost:3000/reset-password?token=%s", token)

	body := fmt.Sprintf(`
		<h2>密碼重設請求</h2>
		<p>我們收到了重設您密碼的請求。請點擊下面的連結來重設密碼：</p>
		<a href="%s">重設密碼</a>
		<p>如果您沒有請求重設密碼，請忽略此郵件。</p>
		<p>此連結將在1小時後過期。</p>
	`, resetURL)

	m.SetBody("text/html", body)

	// 在開發環境中，我們只是記錄郵件內容而不實際發送
	if e.config.Env == "development" {
		fmt.Printf("Password reset email would be sent to %s with reset URL: %s\n", email, resetURL)
		return nil
	}

	return e.dialer.DialAndSend(m)
}

// SendEmail 發送通用郵件
func (e *EmailService) SendEmail(to, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", "noreply@tennis-platform.com")
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	// 在開發環境中，我們只是記錄郵件內容而不實際發送
	if e.config.Env == "development" {
		fmt.Printf("Email would be sent to %s with subject: %s\n", to, subject)
		fmt.Printf("Body: %s\n", body)
		return nil
	}

	return e.dialer.DialAndSend(m)
}

// GenerateToken 生成隨機令牌
func (e *EmailService) GenerateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
