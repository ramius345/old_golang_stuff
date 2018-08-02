package main

import (
	"log"
	"net/smtp"
)

func main() {
	from := "imagesharing392@gmail.com"
	pass := "aybabtu1"
	to := "ramius345@gmail.com"
	body := "http://redgrape:30001"

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: Hello there\n\n" +
		body

	err := smtp.SendMail("smtp.gmail.com:587",
		smtp.PlainAuth("", from, pass, "smtp.gmail.com"),
		from, []string{to}, []byte(msg))

	if err != nil {
		log.Printf("smtp error: %s", err)
		return
	}

}
