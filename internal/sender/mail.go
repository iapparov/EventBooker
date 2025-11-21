package sender

import (
	"eventbooker/internal/config"
	"fmt"
	"net/smtp"
)

type EmailChannel struct {
	smtpHost  string
	smtpPort  int
	smtpEmail string
	smtp      string
}

func NewEmailChannel(cfg *config.AppConfig) *EmailChannel {
	return &EmailChannel{
		smtpHost:  cfg.MailConfig.SMTPHost,
		smtpPort:  cfg.MailConfig.SMTPPort,
		smtpEmail: cfg.MailConfig.SMTPEmail,
		smtp:      cfg.MailConfig.SMTPPassword,
	}
}

func (s *EmailChannel) Send(email string, EventName string, Persons int) error {
	auth := smtp.PlainAuth("", s.smtpEmail, s.smtp, s.smtpHost)
	to := []string{email}
	msg := []byte("To: " + email + "\r\n" +
		"Subject: Booking Cancelation" + "\r\n" +
		"\r\n" +
		fmt.Sprintf("Your booking on %d persons on event: %s just cancelled", Persons, EventName) + "\r\n")
	addr := s.smtpHost + ":" + fmt.Sprint(s.smtpPort)
	err := smtp.SendMail(addr, auth, s.smtpEmail, to, msg)
	if err != nil {
		return err
	}
	return nil
}