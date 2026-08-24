package mailer

import (
	"net/smtp"
	"strings"
)

type Mailer struct {
	host string
	port string
	from string
}

func New(host, port, from string) *Mailer {
	return &Mailer{host: host, port: port, from: from}
}

// ponytail: kirim inline; pindah ke queue (asynq) kalau volume email naik.
func (m *Mailer) Send(to, subject, body string) error {
	addr := m.host + ":" + m.port
	msg := strings.Join([]string{
		"From: " + m.from,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")
	return smtp.SendMail(addr, nil, m.from, []string{to}, []byte(msg))
}
