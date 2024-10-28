package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/BlackSound1/Go-B-and-B/internal/models"
	mail "github.com/xhit/go-simple-mail"
)

func listenForMail() {

	// Create function that runs forever in the background.
	// It's anonymous and an asynchronous goroutine
	go func() {
		for {
			message := <-app.MailChan
			sendMessage(message)
		}
	}()
}

func sendMessage(m models.MailData) {
	// Create mail server object
	server := mail.NewSMTPClient()

	server.Host = "localhost"
	server.Port = 1025
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	// Try to connect
	client, err := server.Connect()
	if err != nil {
		log.Println(err)
	}

	// Create email message
	email := mail.NewMSG()
	email.SetFrom(m.From).AddTo(m.To).SetSubject(m.Subject)
	if m.Template == "" {
		email.SetBody(mail.TextHTML, m.Content)
	} else {
		// Read the HTML template from file
		data, err := os.ReadFile(fmt.Sprintf("./email_templates/%s", m.Template))
		if err != nil {
			log.Println(err)
		}

		// Convert the template to a string
		mailTemplate := string(data)

		// Replace the template variables and send
		msgToSend := strings.Replace(mailTemplate, "[%CONTENT%]", m.Content, 1)
		email.SetBody(mail.TextHTML, msgToSend)
	}

	// Send email
	err = email.Send(client)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Email Sent!")
	}
}
