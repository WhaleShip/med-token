package service

import (
	"gopkg.in/mail.v2"
)

type smtpEmailService struct {
	host     string
	port     int
	username string
	password string
}

func NewSMTPEmailService(host string, port int, user, pass string) *smtpEmailService {
	return &smtpEmailService{host, port, user, pass}
}

func (s *smtpEmailService) Send(to []string, subject, body string) error {
	m := mail.NewMessage()
	m.SetHeader("From", s.username)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	d := mail.NewDialer(s.host, s.port, s.username, s.password)
	return d.DialAndSend(m)
}
