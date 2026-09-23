package service

import (
	"fmt"
	"net/smtp"
	"os"
)

type EmailService struct {
	smtpHost string
	smtpPort string
	username string
	password string
}

func NewEmailService() *EmailService {
	return &EmailService{
		smtpHost: os.Getenv("SMTP_HOST"),
		smtpPort: os.Getenv("SMTP_PORT"),
		username: os.Getenv("SMTP_USERNAME"),
		password: os.Getenv("SMTP_PASSWORD"),
	}
}

func (e *EmailService) SendEmail(
	to string,
	subject string,
	body string,
) error {

	auth := smtp.PlainAuth(
		"",
		e.username,
		e.password,
		e.smtpHost,
	)

	message := []byte(
		"From: " + e.username + "\r\n" +
			"To: " + to + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/plain; charset=UTF-8\r\n" +
			"\r\n" +
			body,
	)

	err := smtp.SendMail(
		e.smtpHost+":"+e.smtpPort,
		auth,
		e.username,
		[]string{to},
		message,
	)

	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}