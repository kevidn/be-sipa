package utils

import (
	"bytes"
	"fmt"
	"net/smtp"
	"os"
	"text/template"
)

type MailData struct {
	NamaLengkap string
	ResetLink   string
	NomorSurat  string
	JenisSurat  string
	Status      string
	Catatan     string
}

func SendStatusUpdateEmail(toEmail string, data MailData) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	from := fmt.Sprintf("SIPA UNESA <%s>", smtpUser)

	subject := fmt.Sprintf("Update Status Pengajuan [%s] - %s", data.NomorSurat, data.Status)

	htmlTemplate := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: sans-serif; background-color: #f4f3ee; color: #1a2a24; }
        .container { max-width: 600px; margin: 40px auto; background: #ffffff; border-radius: 16px; padding: 40px; box-shadow: 0 4px 12px rgba(0,0,0,0.05); }
        .header { border-bottom: 2px solid #5a7a6e; padding-bottom: 20px; margin-bottom: 30px; }
        .status-badge { display: inline-block; padding: 6px 12px; border-radius: 8px; font-weight: bold; background: #e8f5e9; color: #2e7d32; }
        .footer { margin-top: 40px; font-size: 12px; color: #8a9e96; text-align: center; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2 style="color: #5a7a6e; margin: 0;">Update Status Pengajuan</h2>
        </div>
        <p>Halo, <strong>{{.NamaLengkap}}</strong>,</p>
        <p>Status pengajuan surat Anda telah diperbarui:</p>
        
        <div style="background: #f9f8f4; padding: 20px; border-radius: 12px; margin: 20px 0;">
            <p style="margin: 5px 0;"><strong>Nomor:</strong> {{.NomorSurat}}</p>
            <p style="margin: 5px 0;"><strong>Jenis:</strong> {{.JenisSurat}}</p>
            <p style="margin: 5px 0;"><strong>Status Terkini:</strong> <span class="status-badge">{{.Status}}</span></p>
            {{if .Catatan}}
            <p style="margin: 15px 0 5px 0;"><strong>Catatan dari Petugas:</strong></p>
            <p style="font-style: italic; color: #c0392b;">"{{.Catatan}}"</p>
            {{end}}
        </div>

        <p>Silakan pantau perkembangan pengajuan Anda melalui dashboard SIPA.</p>
        
        <div class="footer">
            &copy; 2024 UNIVERSITAS NEGERI SURABAYA<br>
            Sistem Informasi Pelayanan Akademik
        </div>
    </div>
</body>
</html>
`

	t, err := template.New("statusMail").Parse(htmlTemplate)
	if err != nil {
		return err
	}

	var body bytes.Buffer
	mimeHeaders := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body.Write([]byte(fmt.Sprintf("From: %s\nSubject: %s\n%s", from, subject, mimeHeaders)))

	err = t.Execute(&body, data)
	if err != nil {
		return err
	}

	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	err = smtp.SendMail(smtpHost+":"+smtpPort, auth, smtpUser, []string{toEmail}, body.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func SendResetPasswordEmail(toEmail string, data MailData) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	from := fmt.Sprintf("SIPA UNESA <%s>", smtpUser)

	subject := "Reset Kata Sandi - SIPA UNESA"
	
	// HTML Template
	htmlTemplate := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { font-family: 'Inter', Helvetica, Arial, sans-serif; background-color: #f4f3ee; color: #1a2a24; margin: 0; padding: 0; }
        .container { max-width: 600px; margin: 40px auto; background: #ffffff; border-radius: 16px; overflow: hidden; box-shadow: 0 4px 12px rgba(0,0,0,0.05); }
        .header { background: #5a7a6e; padding: 30px; text-align: center; color: white; }
        .header h1 { margin: 0; font-size: 24px; letter-spacing: 1px; }
        .content { padding: 40px; line-height: 1.6; }
        .content p { margin-bottom: 20px; font-size: 16px; color: #4a5e57; }
        .button-container { text-align: center; margin: 30px 0; }
        .button { background-color: #00a86b; color: white; padding: 14px 30px; text-decoration: none; border-radius: 10px; font-weight: 600; display: inline-block; transition: background 0.3s; }
        .footer { padding: 20px; text-align: center; font-size: 12px; color: #8a9e96; background: #f9f8f4; }
        .expiry-note { font-size: 14px; color: #c0392b; font-weight: 500; margin-top: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>SIPA UNESA</h1>
        </div>
        <div class="content">
            <p>Halo, <strong>{{.NamaLengkap}}</strong>,</p>
            <p>Kami menerima permintaan untuk mengatur ulang kata sandi akun SIPA Anda. Klik tombol di bawah ini untuk melanjutkan:</p>
            
            <div class="button-container">
                <a href="{{.ResetLink}}" class="button">Atur Ulang Kata Sandi</a>
            </div>

            <p>Jika Anda tidak merasa melakukan permintaan ini, silakan abaikan email ini. Keamanan akun Anda tetap terjaga.</p>
            
            <p class="expiry-note">Tautan ini hanya berlaku selama 1 jam.</p>
        </div>
        <div class="footer">
            &copy; 2024 UNIVERSITAS NEGERI SURABAYA<br>
            Sistem Informasi Pelayanan Akademik digital
        </div>
    </div>
</body>
</html>
`

	t, err := template.New("mail").Parse(htmlTemplate)
	if err != nil {
		return err
	}

	var body bytes.Buffer
	mimeHeaders := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body.Write([]byte(fmt.Sprintf("From: %s\nSubject: %s\n%s", from, subject, mimeHeaders)))

	err = t.Execute(&body, data)
	if err != nil {
		return err
	}

	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	err = smtp.SendMail(smtpHost+":"+smtpPort, auth, smtpUser, []string{toEmail}, body.Bytes())
	if err != nil {
		return err
	}

	return nil
}
