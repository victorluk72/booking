package main

import (
	"fmt"
	"time"

	"github.com/victorluk72/booking/internal/models"
	mail "github.com/xhit/go-simple-mail/v2"
)

func listenForMail() {

	//Run an anynimouse function asyncronically (use go routine)
	//Listen all the time for incoming data
	go func() {

		// use for loop with no conditions (make it runs infinetely)
		//Create a messge that was sent from channel
		msg := <-app.MailChan
		sendMsg(msg) //sending email by using custom buit function sendMsg() see below

	}()

}

// sendMsg sends email messages, use in async function above
// Get all parameters for function from struct MAilData
func sendMsg(m models.MailData) {

	//Define mail SERVER parameters
	server := mail.NewSMTPClient()
	server.Host = "smtp.mailtrap.io"
	server.Port = 2525
	server.Username = "209f2634471d77"
	server.Password = "7515969fa1abd7"
	server.KeepAlive = false
	server.ConnectTimeout = 10 + time.Second
	server.SendTimeout = 10 + time.Second

	//Define mail CLIENT parameters
	client, err := server.Connect()
	if err != nil {
		errorLog.Println(err)
	}

	//construct our email MESSAGE
	//create new empty message and then set main email parameters (from, to, subject)
	email := mail.NewMSG()
	email.SetFrom(m.From).AddTo(m.To).SetSubject(m.Subject)
	email.SetBody(mail.TextHTML, m.Content)

	//Send email now!
	err = email.Send(client)
	if err != nil {
		errorLog.Println(err)
	} else {
		fmt.Println("Email sent....")
	}

}
