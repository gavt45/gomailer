package main

import (
	"log"
	"net/smtp"
)

func Send(to string, subject string, body string, uid string, config *Config) {
	from := config.Email
	pass := config.Password

	//to := "foobarbazz@mailinator.com"

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: "+subject+"\n\n" +
		string(body)

	err := smtp.SendMail(config.SmtpAddr+":587",
		smtp.PlainAuth("", from, pass, config.SmtpAddr),
		from, []string{to}, []byte(msg))

	if err != nil {
		log.Println("smtp error: ", err)
		var res = TaskResult{
			Code:    1,
			Message: err.Error(),
		}
		Tasks.Lock()
		Tasks.m[uid] = res
		Tasks.Unlock()
		return
	}
	log.Println("Send mail to",to,"with subject",subject,"successful",uid)
	var res = TaskResult{
		Code:    0,
		Message: "OK",
	}
	Tasks.Lock()
	Tasks.m[uid] = res
	Tasks.Unlock()
	//log.Print("sent, visit http://foobarbazz.mailinator.com")
}
