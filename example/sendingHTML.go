package main

import (
	"github.com/supme/directEmail"
	"log"
)

func main() {
	email := directEmail.New()
	email.From = "sender@example.com"
	email.Data.From = "Sender <sender@example.com>"
	email.Data.Subject = "Test email"
	email.Data.Html = "<h2>Привет!</h2><br/>"

	email.To = "user@example.com"
	email.Data.To = "User <user@example.com>"
	email.Data.Render()
	err := email.Send()
	if err != nil {
		log.Printf("Send email with error '%s'", err.Error())
	}

}

