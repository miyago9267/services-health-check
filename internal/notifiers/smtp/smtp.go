package smtp

import (
	"context"
	"crypto/tls"
	"fmt"
	"mime"
	"net"
	"net/smtp"
	"strings"
	"time"

	"services-health-check/internal/core/notify"
	"services-health-check/internal/notifiers/format"
)

type Notifier struct {
	NameValue     string
	Host          string
	Port          int
	Username      string
	Password      string
	From          string
	To            []string
	Subject       string
	Timeout       time.Duration
	ImplicitTLS   bool
	SkipVerifyTLS bool
}

func (n *Notifier) Name() string {
	return n.NameValue
}

func (n *Notifier) Send(ctx context.Context, event notify.Event) error {
	if n.Host == "" {
		return fmt.Errorf("smtp host is required")
	}
	if n.Port == 0 {
		n.Port = 587
	}
	if n.From == "" {
		return fmt.Errorf("smtp from is required")
	}
	if len(n.To) == 0 {
		return fmt.Errorf("smtp to is required")
	}

	subject := n.Subject
	if subject == "" {
		subject = fmt.Sprintf("[%s] %s", event.Status, event.Summary)
	}
	subject = mime.QEncoding.Encode("utf-8", subject)

	details := format.DetailsList(event.Details)
	bodyLines := []string{
		fmt.Sprintf("[%s] %s", event.Status, event.Summary),
		fmt.Sprintf("服務: %s", event.Service),
		fmt.Sprintf("狀態: %s", event.Status),
		fmt.Sprintf("時間: %s", event.OccurredAt.Format(time.RFC3339)),
		"",
		"細節:",
		details,
	}
	body := strings.Join(bodyLines, "\n")

	msg := buildMessage(n.From, n.To, subject, body)
	addr := fmt.Sprintf("%s:%d", n.Host, n.Port)

	client, err := n.dialSMTP(ctx, addr)
	if err != nil {
		return err
	}
	defer client.Close()

	if n.Username != "" {
		auth := smtp.PlainAuth("", n.Username, n.Password, n.Host)
		if err := client.Auth(auth); err != nil {
			return err
		}
	}
	if err := client.Mail(n.From); err != nil {
		return err
	}
	for _, to := range n.To {
		if err := client.Rcpt(to); err != nil {
			return err
		}
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write([]byte(msg)); err != nil {
		_ = w.Close()
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return client.Quit()
}

func (n *Notifier) dialSMTP(ctx context.Context, addr string) (*smtp.Client, error) {
	timeout := n.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	tlsConfig := &tls.Config{
		ServerName:         n.Host,
		InsecureSkipVerify: n.SkipVerifyTLS,
	}

	if n.ImplicitTLS {
		conn, err := tls.DialWithDialer(&net.Dialer{Timeout: timeout}, "tcp", addr, tlsConfig)
		if err != nil {
			return nil, err
		}
		return smtp.NewClient(conn, n.Host)
	}

	conn, err := (&net.Dialer{Timeout: timeout}).DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}
	client, err := smtp.NewClient(conn, n.Host)
	if err != nil {
		return nil, err
	}
	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(tlsConfig); err != nil {
			_ = client.Close()
			return nil, err
		}
	}
	return client, nil
}

func buildMessage(from string, to []string, subject string, body string) string {
	headers := []string{
		"From: " + from,
		"To: " + strings.Join(to, ", "),
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
	}
	return strings.Join(headers, "\r\n") + "\r\n\r\n" + body
}
