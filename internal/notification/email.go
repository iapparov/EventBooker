package notification

import (
	"fmt"
	"net/smtp"

	"eventbooker/internal/config"
)

// EmailSender sends cancellation notifications via email.
type EmailSender struct {
	smtpHost     string
	smtpPort     int
	smtpEmail    string
	smtpPassword string
}

// NewEmailSender creates a new EmailSender.
func NewEmailSender(cfg *config.AppConfig) *EmailSender {
	return &EmailSender{
		smtpHost:     cfg.Mail.SMTPHost,
		smtpPort:     cfg.Mail.SMTPPort,
		smtpEmail:    cfg.Mail.SMTPEmail,
		smtpPassword: cfg.Mail.SMTPPassword,
	}
}

// Send sends a booking cancellation email.
func (s *EmailSender) Send(email, eventName string, persons int) error {
	auth := smtp.PlainAuth("", s.smtpEmail, s.smtpPassword, s.smtpHost)
	to := []string{email}
	msg := []byte("To: " + email + "\r\n" +
		"Subject: Booking Cancelation\r\n" +
		"\r\n" +
		fmt.Sprintf("Your booking on %d persons on event: %s just cancelled\r\n", persons, eventName))
	addr := fmt.Sprintf("%s:%d", s.smtpHost, s.smtpPort)
	return smtp.SendMail(addr, auth, s.smtpEmail, to, msg)
}
