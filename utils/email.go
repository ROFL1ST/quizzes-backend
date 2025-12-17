package utils

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/gomail.v2"
)

// Fungsi Dasar Pengiriman Email
func sendMail(to string, subject string, htmlBody string) error {
	host := os.Getenv("SMTP_HOST")
	portStr := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	password := os.Getenv("SMTP_PASS")

	// Konversi port ke integer
	port, _ := strconv.Atoi(portStr)

	m := gomail.NewMessage()
	m.SetHeader("From", user)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", htmlBody)

	// Setup Dialer untuk Gmail
	d := gomail.NewDialer(host, port, user, password)

	// Kirim
	if err := d.DialAndSend(m); err != nil {
		return err
	}
	return nil
}

// 1. Template Email Verifikasi (Dipanggil saat tambah email)
func SendVerificationEmail(toEmail, token string) error {
	frontendURL := os.Getenv("FRONTEND_URL")
	link := fmt.Sprintf("%s/verify-email?token=%s", frontendURL, token)

	subject := "Verifikasi Email Akun Quiz App"
	body := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; padding: 20px;">
			<h2>Halo!</h2>
			<p>Anda baru saja menambahkan email ini ke akun Quiz App.</p>
			<p>Klik tombol di bawah untuk memverifikasi email Anda:</p>
			<a href="%s" style="background-color: #4CAF50; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px; display: inline-block;">Verifikasi Email</a>
			<br><br>
			<p>Atau copy link ini: <br> <a href="%s">%s</a></p>
			<p><i>Abaikan email ini jika Anda tidak merasa mendaftarkannya.</i></p>
		</div>
	`, link, link, link)

	return sendMail(toEmail, subject, body)
}

// 2. Template Email Reset Password (Dipanggil saat lupa password)
func SendResetPasswordEmail(toEmail, token string) error {
	frontendURL := os.Getenv("FRONTEND_URL")
	link := fmt.Sprintf("%s/reset-password?token=%s", frontendURL, token)

	subject := "Permintaan Reset Password"
	body := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; padding: 20px;">
			<h2>Reset Password</h2>
			<p>Kami menerima permintaan untuk mereset password Anda.</p>
			<p>Klik tombol di bawah untuk membuat password baru:</p>
			<a href="%s" style="background-color: #d9534f; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px; display: inline-block;">Reset Password</a>
			<br><br>
			<p>Link ini hanya berlaku selama 1 jam.</p>
			<p><i>Jika bukan Anda yang meminta, abaikan saja email ini.</i></p>
		</div>
	`, link)

	return sendMail(toEmail, subject, body)
}