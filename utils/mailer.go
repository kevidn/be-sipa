package utils

import (
	"bytes"
	"fmt"
	"io"
	"net/smtp"
	"text/template"

	"github.com/kevidn/be-sipa/config"
	"github.com/kevidn/be-sipa/models"
	"gopkg.in/gomail.v2"
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
	var setting models.SystemSetting
	if err := config.DB.First(&setting, 1).Error; err != nil {
		return fmt.Errorf("gagal memuat pengaturan sistem: %v", err)
	}

	if !setting.EmailNotification {
		return nil // Notifikasi email dimatikan, tidak perlu error
	}

	smtpHost := setting.SMTPServer
	smtpPort := fmt.Sprintf("%d", setting.SMTPPort)
	smtpUser := setting.SMTPUsername
	smtpPass := setting.SMTPPassword
	
	if smtpHost == "" || smtpUser == "" {
		return fmt.Errorf("konfigurasi SMTP belum diatur di sistem")
	}

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

// SendEmailWithKitir sends a status update email and attaches a PDF file from bytes.
func SendEmailWithKitir(toEmail string, data MailData, pdfAttachment []byte) error {
	var setting models.SystemSetting
	if err := config.DB.First(&setting, 1).Error; err != nil {
		return fmt.Errorf("gagal memuat pengaturan sistem: %v", err)
	}

	if !setting.EmailNotification {
		return nil
	}

	if setting.SMTPServer == "" || setting.SMTPUsername == "" {
		return fmt.Errorf("konfigurasi SMTP belum diatur di sistem")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", fmt.Sprintf("SIPA UNESA <%s>", setting.SMTPUsername))
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", fmt.Sprintf("Bukti Pengajuan [%s] - SIPA UNESA", data.NomorSurat))

	htmlTemplate := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: sans-serif; background-color: #f4f3ee; color: #1a2a24; }
        .container { max-width: 600px; margin: 40px auto; background: #ffffff; border-radius: 16px; padding: 40px; box-shadow: 0 4px 12px rgba(0,0,0,0.05); }
        .header { border-bottom: 2px solid #5a7a6e; padding-bottom: 20px; margin-bottom: 30px; }
        .footer { margin-top: 40px; font-size: 12px; color: #8a9e96; text-align: center; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2 style="color: #5a7a6e; margin: 0;">Pengajuan Surat Berhasil</h2>
        </div>
        <p>Halo, <strong>{{.NamaLengkap}}</strong>,</p>
        <p>Pengajuan surat Anda telah kami terima dengan detail berikut:</p>
        
        <div style="background: #f9f8f4; padding: 20px; border-radius: 12px; margin: 20px 0;">
            <p style="margin: 5px 0;"><strong>Nomor:</strong> {{.NomorSurat}}</p>
            <p style="margin: 5px 0;"><strong>Jenis:</strong> {{.JenisSurat}}</p>
        </div>

        <p>Terlampir adalah bukti pengajuan digital (Kitir) Anda dalam bentuk PDF. Harap simpan bukti ini dengan baik.</p>
        
        <div class="footer">
            &copy; 2024 UNIVERSITAS NEGERI SURABAYA<br>
            Sistem Informasi Pelayanan Akademik
        </div>
    </div>
</body>
</html>
`
	t, err := template.New("kitirMail").Parse(htmlTemplate)
	if err != nil {
		return err
	}

	var body bytes.Buffer
	err = t.Execute(&body, data)
	if err != nil {
		return err
	}

	m.SetBody("text/html", body.String())

	// Attach PDF from memory
	m.Attach(fmt.Sprintf("Kitir_%s.pdf", data.NomorSurat), gomail.SetCopyFunc(func(w io.Writer) error {
		_, err := w.Write(pdfAttachment)
		return err
	}))

	d := gomail.NewDialer(setting.SMTPServer, setting.SMTPPort, setting.SMTPUsername, setting.SMTPPassword)

	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}

func SendResetPasswordEmail(toEmail string, data MailData) error {
	var setting models.SystemSetting
	if err := config.DB.First(&setting, 1).Error; err != nil {
		return fmt.Errorf("gagal memuat pengaturan sistem: %v", err)
	}

	// Email reset password tetap dikirim walau notifikasi dimatikan, karena penting.
	
	smtpHost := setting.SMTPServer
	smtpPort := fmt.Sprintf("%d", setting.SMTPPort)
	smtpUser := setting.SMTPUsername
	smtpPass := setting.SMTPPassword
	
	if smtpHost == "" || smtpUser == "" {
		return fmt.Errorf("konfigurasi SMTP belum diatur di sistem")
	}

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

func SendSLAWarningEmail(toEmail string, namaTendik string, nomorSurat string, sisaHari int) error {
	var setting models.SystemSetting
	if err := config.DB.First(&setting, 1).Error; err != nil {
		return err
	}
	if !setting.EmailNotification {
		return nil
	}
	m := gomail.NewMessage()
	m.SetHeader("From", fmt.Sprintf("SIPA UNESA <%s>", setting.SMTPUsername))
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", fmt.Sprintf("[Penting] Peringatan SLA Surat %s", nomorSurat))

	htmlContent := fmt.Sprintf(`
		<p>Halo <b>%s</b>,</p>
		<p>Terdapat pengajuan surat yang akan segera melewati batas waktu (SLA):</p>
		<p>Nomor Surat: <b>%s</b></p>
		<p>Sisa waktu efektif: <b>%d jam/hari</b></p>
		<p>Mohon segera memproses surat tersebut agar tidak terjadi keterlambatan pelayanan.</p>
	`, namaTendik, nomorSurat, sisaHari)
	m.SetBody("text/html", htmlContent)
	d := gomail.NewDialer(setting.SMTPServer, setting.SMTPPort, setting.SMTPUsername, setting.SMTPPassword)
	return d.DialAndSend(m)
}

func SendSLAEscalationEmail(toEmail string, namaKaprodi string, nomorSurat string, namaTendik string) error {
	var setting models.SystemSetting
	if err := config.DB.First(&setting, 1).Error; err != nil {
		return err
	}
	if !setting.EmailNotification {
		return nil
	}
	m := gomail.NewMessage()
	m.SetHeader("From", fmt.Sprintf("SIPA UNESA <%s>", setting.SMTPUsername))
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", fmt.Sprintf("[Eskalasi] Pelanggaran SLA Surat %s", nomorSurat))

	htmlContent := fmt.Sprintf(`
		<p>Yth. <b>%s</b>,</p>
		<p>Sistem mendeteksi bahwa pengajuan surat berikut telah melebihi batas waktu pelayanan yang ditetapkan (SLA Terlampaui):</p>
		<ul>
			<li>Nomor Surat: <b>%s</b></li>
			<li>Diproses oleh: <b>%s</b></li>
		</ul>
		<p>Mohon tindak lanjut dari Anda selaku Kepala Program Studi.</p>
	`, namaKaprodi, nomorSurat, namaTendik)
	m.SetBody("text/html", htmlContent)
	d := gomail.NewDialer(setting.SMTPServer, setting.SMTPPort, setting.SMTPUsername, setting.SMTPPassword)
	return d.DialAndSend(m)
}
