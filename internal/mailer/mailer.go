// Package mailer delivers one-time codes.
//
// SMTP rather than a provider SDK: a self-hosted deployment must be able to
// point at any mail service, and every provider speaks SMTP.
package mailer

import (
	"fmt"
	"log/slog"
	"net/smtp"
	"strings"
)

// Mailer sends a message.
type Mailer interface {
	SendCode(to, code string) error
}

// Stdout prints codes to the log instead of sending them.
//
// Development only. It is the default when no SMTP host is configured, and it
// logs loudly, because a production deployment that quietly falls back to this
// is one where every sign-in code is in the log file.
type Stdout struct{ ServiceName string }

func (s *Stdout) SendCode(to, code string) error {
	slog.Warn("email not configured; printing the sign-in code instead",
		"to", to, "code", code)
	return nil
}

// SMTP sends through a real server.
type SMTP struct {
	Host, Username, Password, From, ServiceName string
	Port                                        int
}

func (m *SMTP) SendCode(to, code string) error {
	body := fmt.Sprintf(""+
		"From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: Your %s sign-in code\r\n"+
		"\r\n"+
		"Your sign-in code is %s\r\n\r\n"+
		"It expires in 10 minutes. If you didn't ask for it, you can ignore this.\r\n",
		m.From, to, m.ServiceName, code)

	addr := fmt.Sprintf("%s:%d", m.Host, m.Port)
	auth := smtp.PlainAuth("", m.Username, m.Password, m.Host)
	if err := smtp.SendMail(addr, auth, m.From, []string{to}, []byte(body)); err != nil {
		return fmt.Errorf("mailer: sending to %s: %w", redact(to), err)
	}
	return nil
}

// redact keeps an address out of logs while leaving enough to debug with.
func redact(email string) string {
	at := strings.LastIndex(email, "@")
	if at <= 1 {
		return "***"
	}
	return email[:1] + "***" + email[at:]
}
