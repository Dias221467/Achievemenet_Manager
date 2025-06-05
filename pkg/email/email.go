package email

import (
	"fmt"
	"net/smtp"
	"os"
)

// SendEmail sends a plain text email using SMTP.
func SendEmail(to, subject, body string) error {
	from := os.Getenv("SMTP_SENDER")
	password := os.Getenv("SMTP_PASSWORD")
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")

	auth := smtp.PlainAuth("", from, password, smtpHost)

	msg := []byte("To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" + body + "\r\n")

	address := smtpHost + ":" + smtpPort

	err := smtp.SendMail(address, auth, from, []string{to}, msg)
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}
	return nil
}
