package main

type Config struct {
	Port		int 	`json:"port"`
	ServerCert	string	`json:"server_cert"`
	ServerKey	string	`json:"server_key"`
	Secret		string	`json:"secret"`
	Email		string	`json:"email"`
	SmtpAddr	string	`json:"smtp_addr"`
	Password	string	`json:"password"`
}
