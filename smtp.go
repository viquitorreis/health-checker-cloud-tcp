package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"os"
	"text/template"
)

type EmailRequest struct {
	From    string
	To      []string
	Subject string
	Body    string
}

func NewEmailRequest(to []string, subject string, body string, from ...string) *EmailRequest {
	if len(from) == 0 {
		return &EmailRequest{
			From:    os.Getenv("SMTP_EMAIL_FROM"),
			To:      to,
			Subject: subject,
			Body:    body,
		}
	}

	return &EmailRequest{
		From:    from[0],
		To:      to,
		Subject: subject,
		Body:    body,
	}
}

func (r *EmailRequest) SendEmail() error {
	smtpPassword := os.Getenv("SMTP_EMAIL_PASSWORD")
	smtpUser := os.Getenv("SMTP_EMAIL_USERNAME")
	smtpHost := os.Getenv("SMTP_EMAIL_HOST")
	smtpPort := os.Getenv("SMTP_EMAIL_PORT")

	auth := smtp.PlainAuth("", smtpUser, smtpPassword, smtpHost)

	r.Body = "MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n" +
		"Subject: " + r.Subject + "\r\n\r\n" +
		r.Body

	conn, err := net.Dial("tcp", smtpHost+":"+smtpPort)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}

	client, err := smtp.NewClient(conn, smtpHost)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}

	tlsConfig := &tls.Config{
		ServerName: smtpHost,
	}

	if err = client.StartTLS(tlsConfig); err != nil {
		return err
	}

	if err = client.Auth(auth); err != nil {
		return err
	}

	if err := client.Mail(r.From); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	for _, to := range r.To {
		if err := client.Rcpt(to); err != nil {
			return fmt.Errorf("failed to set recipient: %w", err)
		}
	}

	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to open data connection: %w", err)
	}

	_, err = wc.Write([]byte(r.Body))
	if err != nil {
		return fmt.Errorf("failed to write email body: %w", err)
	}
	err = wc.Close()
	if err != nil {
		return fmt.Errorf("failed to close data connection: %w", err)
	}

	if err := client.Quit(); err != nil {
		return fmt.Errorf("failed to quit: %w", err)
	}

	return nil
}

func (r *EmailRequest) ParseTemplate(templatePath string, data any) error {
	templateContent, err := TemplateEmbedded.ReadFile(templatePath)
	if err != nil {
		return err
	}

	tmpl, err := template.New("template").Parse(string(templateContent))
	if err != nil {
		return err
	}

	var tpl bytes.Buffer
	err = tmpl.Execute(&tpl, data)
	if err != nil {
		return err
	}

	r.Body = tpl.String()

	return nil
}
