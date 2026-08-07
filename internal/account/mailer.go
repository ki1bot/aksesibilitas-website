package account

import (
	"fmt"
	"html"
	"net"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

func (handler *Handler) emailConfigured() bool {
	return handler.cfg.SMTPHost != "" &&
		handler.cfg.SMTPPort > 0 &&
		handler.cfg.SMTPFromEmail != ""
}

func (handler *Handler) sendResetEmail(
	user resetUser,
	resetURL string,
) error {
	address := net.JoinHostPort(
		handler.cfg.SMTPHost,
		strconv.Itoa(handler.cfg.SMTPPort),
	)

	var smtpAuth smtp.Auth

	if handler.cfg.SMTPUsername != "" {
		smtpAuth = smtp.PlainAuth(
			"",
			handler.cfg.SMTPUsername,
			handler.cfg.SMTPPassword,
			handler.cfg.SMTPHost,
		)
	}

	subject :=
		"Buat password baru untuk akun AksesCheck"

	body := fmt.Sprintf(
		`<!doctype html>
<html lang="id">
  <body style="font-family:Arial,sans-serif;color:#1c1815;line-height:1.6">
    <h1 style="font-size:22px">Buat password baru</h1>
    <p>Halo %s,</p>
    <p>Kami menerima permintaan untuk mengganti password akun AksesCheck Anda.</p>
    <p><a href="%s" style="display:inline-block;padding:12px 18px;border-radius:10px;background:#0f766e;color:#ffffff;text-decoration:none;font-weight:700">Buat password baru</a></p>
    <p>Tautan ini berlaku selama %s dan hanya dapat digunakan satu kali.</p>
    <p>Jika Anda tidak meminta perubahan ini, abaikan email ini.</p>
  </body>
</html>`,
		html.EscapeString(user.Name),
		html.EscapeString(resetURL),
		humanDuration(
			handler.cfg.PasswordResetTTL,
		),
	)

	message := strings.Join(
		[]string{
			"From: " +
				handler.cfg.SMTPFromName +
				" <" +
				handler.cfg.SMTPFromEmail +
				">",
			"To: " + user.Email,
			"Subject: " + subject,
			"MIME-Version: 1.0",
			"Content-Type: text/html; charset=UTF-8",
			"",
			body,
		},
		"\r\n",
	)

	return smtp.SendMail(
		address,
		smtpAuth,
		handler.cfg.SMTPFromEmail,
		[]string{user.Email},
		[]byte(message),
	)
}

func humanDuration(
	duration time.Duration,
) string {
	if duration%time.Hour == 0 {
		hours := int(duration / time.Hour)

		return strconv.Itoa(hours) + " jam"
	}

	minutes := int(duration / time.Minute)

	return strconv.Itoa(minutes) + " menit"
}
