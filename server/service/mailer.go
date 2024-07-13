package service

import (
	"fmt"
	"net/smtp"

	"github.com/spf13/viper"
	"github.com/vakshit-zomato/assignment-server/domain"
)

const EmailSubject = "Notification"

func SendEmailBulk(notifs []domain.Notification) error {
	from := viper.GetString("email.from")
	password := viper.GetString("email.password")
	smtpHost := viper.GetString("email.smtp_host")
	smtpPort := viper.GetString("email.smtp_port")

	if from == "" || password == "" || smtpHost == "" || smtpPort == "" {
		return fmt.Errorf("email credentials and SMTP configuration must be set")
	}

	// Set up authentication information.
	auth := smtp.PlainAuth("", from, password, smtpHost)
	for _, notif := range notifs {
		// Create the email message.
		message := []byte(fmt.Sprintf("Subject: %s\r\n\r\n%s", EmailSubject, notif.Message))

		// Send the email.
		err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{notif.Email}, message)
		if err != nil {
			return err
		}
	}
	return nil
}
