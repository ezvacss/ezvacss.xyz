package handlers

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
)

// SendDetailedReportEmail sends an automated email notification when a link is reported.
func SendDetailedReportEmail(shortCode, reason, reporterName, reporterEmail, reportType string) {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASSWORD")
	adminEmail := os.Getenv("ADMIN_EMAIL")
	senderEmail := os.Getenv("SENDER_EMAIL")

	// Skip if SMTP is not fully configured
	if smtpHost == "" || smtpPort == "" || smtpUser == "" || smtpPass == "" || adminEmail == "" {
		log.Println("SMTP notification skipped: environment variables are not fully configured.")
		return
	}

	if senderEmail == "" {
		senderEmail = smtpUser
	}

	if reporterName == "" {
		reporterName = "Anonymous"
	}
	if reporterEmail == "" {
		reporterEmail = "Not provided"
	}

	subject := fmt.Sprintf("Subject: [Abuse Report] Link /%s Reported (%s)\n", shortCode, reportType)
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body := fmt.Sprintf(`
		<h2>Abuse Report Submitted</h2>
		<p><strong>Short Code:</strong> <a href="https://ezvacss.xyz/%s">/%s</a></p>
		<p><strong>Report Type:</strong> %s</p>
		<p><strong>Reporter Name:</strong> %s</p>
		<p><strong>Reporter Email:</strong> %s</p>
		<p><strong>Reason for report:</strong></p>
		<blockquote style="background: #f9f9f9; border-left: 10px solid #e74c3c; margin: 1.5em 10px; padding: 1rem; font-style: italic;">
			%s
		</blockquote>
		<hr style="border: 0; border-top: 1px solid #eee;">
		<p style="font-size: 0.85rem; color: #888;">This is an automated message from ezvacss.xyz operations.</p>
	`, shortCode, shortCode, reportType, reporterName, reporterEmail, reason)

	msg := []byte(subject + mime + body)
	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)

	go func() {
		addr := smtpHost + ":" + smtpPort
		err := smtp.SendMail(addr, auth, senderEmail, []string{adminEmail}, msg)
		if err != nil {
			log.Printf("Failed to send abuse report email: %v", err)
		} else {
			log.Printf("Abuse report email successfully sent to %s for code /%s", adminEmail, shortCode)
		}
	}()
}
