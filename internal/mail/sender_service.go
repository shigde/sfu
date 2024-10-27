package mail

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"net/url"
)

type SenderService struct {
	cfg         *MailConfig
	instanceUrl *url.URL
}

func NewSenderService(config *MailConfig, instanceUrl *url.URL) *SenderService {
	return &SenderService{
		cfg:         config,
		instanceUrl: instanceUrl,
	}
}

func (s *SenderService) SendActivateAccountMail(name string, email string, link string) error {

	if !s.cfg.Enable {
		return nil
	}

	// Receiver email address.
	to := []string{email}

	// Authentication.
	auth := smtp.PlainAuth("", s.cfg.From, s.cfg.Pass, s.cfg.SmtpHost)

	t, err := template.ParseFiles("internal/mail/template/activate_account.html")
	if err != nil {
		return fmt.Errorf("parsing account email template: %w", err)
	}

	var body bytes.Buffer

	mimeHeaders := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	if _, err = body.Write([]byte(fmt.Sprintf("Subject: Activate Your Account \n%s\n\n", mimeHeaders))); err != nil {
		return fmt.Errorf("writing account email boddy: %w", err)
	}

	t.Execute(&body, struct {
		User     string
		Link     string
		Instance string
	}{
		User:     name,
		Link:     link,
		Instance: s.instanceUrl.String(),
	})

	// Sending email.
	if err = smtp.SendMail(s.cfg.SmtpHost+":"+s.cfg.SmtpPort, auth, s.cfg.From, to, body.Bytes()); err != nil {
		return fmt.Errorf("sending account email boddy: %w", err)
	}
	return nil
}

func (s *SenderService) SendForgotPasswordMail(name string, email string, link string) error {

	if !s.cfg.Enable {
		return nil
	}

	// Receiver email address.
	to := []string{email}

	// Authentication.
	auth := smtp.PlainAuth("", s.cfg.From, s.cfg.Pass, s.cfg.SmtpHost)

	t, err := template.ParseFiles("internal/mail/template/forgot_password.html")
	if err != nil {
		return fmt.Errorf("parsing pass forgot email template: %w", err)
	}

	var body bytes.Buffer

	mimeHeaders := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	if _, err = body.Write([]byte(fmt.Sprintf("Subject: Forgot Password \n%s\n\n", mimeHeaders))); err != nil {
		return fmt.Errorf("writing pass forgot email boddy: %w", err)
	}

	t.Execute(&body, struct {
		User     string
		Link     string
		Instance string
	}{
		User:     name,
		Link:     link,
		Instance: s.instanceUrl.String(),
	})

	// Sending email.
	if err = smtp.SendMail(s.cfg.SmtpHost+":"+s.cfg.SmtpPort, auth, s.cfg.From, to, body.Bytes()); err != nil {
		return fmt.Errorf("sending pass forgot email boddy: %w", err)
	}
	return nil
}
