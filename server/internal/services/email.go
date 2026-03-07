package services

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"

	"birdactyl-panel-backend/internal/config"
)

func SendEmail(to, subject, htmlBody string) error {
	cfg := config.Get()
	if !cfg.SMTPEnabled() {
		return fmt.Errorf("SMTP is not configured")
	}

	from := cfg.SMTP.FromEmail
	fromName := cfg.SMTP.FromName
	host := cfg.SMTP.Host
	port := cfg.SMTP.Port
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))

	headers := fmt.Sprintf("From: %s <%s>\r\n", fromName, from)
	headers += fmt.Sprintf("To: %s\r\n", to)
	headers += fmt.Sprintf("Subject: %s\r\n", subject)
	headers += "MIME-Version: 1.0\r\n"
	headers += "Content-Type: text/html; charset=\"UTF-8\"\r\n"
	headers += "\r\n"

	msg := []byte(headers + htmlBody)

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	tlsConfig := &tls.Config{ServerName: host}
	if err := client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("STARTTLS failed: %w", err)
	}

	if cfg.SMTP.Username != "" {
		auth := smtp.PlainAuth("", cfg.SMTP.Username, cfg.SMTP.Password, host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP auth failed: %w", err)
		}
	}

	if err := client.Mail(from); err != nil {
		return fmt.Errorf("MAIL FROM failed: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("RCPT TO failed: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA failed: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close message: %w", err)
	}

	return client.Quit()
}
