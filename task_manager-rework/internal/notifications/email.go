package notifications

import (
	"fmt"
	"net/smtp"
	"os"
	"strconv"
	"strings"
)

type EmailSender struct {
	host     string
	port     string
	username string
	password string
	from     string
}

func NewEmailSender() *EmailSender {
	return &EmailSender{
		host:     strings.TrimSpace(os.Getenv("SMTP_HOST")),
		port:     envOr("SMTP_PORT", "587"),
		username: os.Getenv("SMTP_USERNAME"),
		password: os.Getenv("SMTP_PASSWORD"),
		from:     strings.TrimSpace(os.Getenv("SMTP_FROM")),
	}
}

func (s *EmailSender) Configured() bool {
	if s.host == "" || s.from == "" {
		return false
	}
	port, err := strconv.Atoi(s.port)
	return err == nil && port > 0 && port < 65536
}

func (s *EmailSender) Send(to, subject, body string) error {
	if !s.Configured() {
		return fmt.Errorf("email delivery is not configured: set SMTP_HOST, SMTP_PORT and SMTP_FROM")
	}
	to = strings.TrimSpace(to)
	if to == "" || strings.ContainsAny(to, "\r\n") || strings.ContainsAny(subject, "\r\n") {
		return fmt.Errorf("invalid email recipient or subject")
	}
	auth := smtp.Auth(nil)
	if s.username != "" {
		auth = smtp.PlainAuth("", s.username, s.password, s.host)
	}
	message := []byte("From: " + s.from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n\r\n" + body + "\r\n")
	return smtp.SendMail(s.host+":"+s.port, auth, s.from, []string{to}, message)
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
