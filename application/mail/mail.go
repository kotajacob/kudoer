// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package mail

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"
	"time"

	"github.com/go-mail/mail/v2"
)

const resetTMPL = "reset.tmpl"

//go:embed "reset.tmpl"
var EFS embed.FS

type Mailer struct {
	dialer *mail.Dialer
	sender string
}

func New(host string, port int, username, password, sender string) *Mailer {
	dialer := mail.NewDialer(host, port, username, password)
	dialer.Timeout = 10 * time.Second

	return &Mailer{
		dialer: dialer,
		sender: sender,
	}
}

// PasswordReset sends a password reset email to a recipient.
func (m *Mailer) PasswordReset(recipient string, data any) error {
	tmpl, err := template.New(resetTMPL).ParseFS(EFS, resetTMPL)
	if err != nil {
		return fmt.Errorf("failed parsing email template: %v", err)
	}

	var b bytes.Buffer
	err = tmpl.Execute(&b, data)
	if err != nil {
		return fmt.Errorf("failed executing email template: %v", err)
	}

	msg := mail.NewMessage()
	msg.SetHeader("To", recipient)
	msg.SetHeader("From", m.sender)
	msg.SetHeader("Subject", "Kudoer - Password Reset")
	msg.SetBody("text/plain", b.String())

	for i := 0; i < 3; i++ {
		err = m.dialer.DialAndSend(msg)
		if err == nil {
			return nil
		}

		time.Sleep(5 * time.Second)
	}

	return err
}
