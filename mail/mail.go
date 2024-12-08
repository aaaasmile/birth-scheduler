package mail

import (
	"birthsch/conf"
	"bytes"
	"crypto/tls"
	"log"
	"net"
	"net/smtp"
)

type MailSender struct {
	relay    conf.Relay
	simulate bool
	emailTo  string
	message  *bytes.Buffer
}

func (ms *MailSender) SendEmailViaRelay() error {
	log.Println("Send email using relay host")

	if ms.simulate {
		log.Println("This is a simulation, e-mail is not sent")
		return nil
	}

	servername := ms.relay.Host

	host, _, _ := net.SplitHostPort(servername)

	auth := smtp.PlainAuth("", ms.relay.User, ms.relay.Secret, host)

	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}

	log.Println("Dial server ", servername)
	conn, err := tls.Dial("tcp", servername, tlsconfig)
	if err != nil {
		return err
	}

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}

	log.Println("Send smtp Auth")
	if err = c.Auth(auth); err != nil {
		return err
	}

	log.Println("send From")
	if err = c.Mail(ms.relay.Mail); err != nil {
		return err
	}
	log.Println("send To")
	if err = c.Rcpt(ms.emailTo); err != nil {
		return err
	}

	w, err := c.Data()
	if err != nil {
		return err
	}
	log.Println("Send the message to the relay")
	_, err = w.Write(ms.message.Bytes())
	if err != nil {
		return err
	}
	log.Println("Close relay")
	err = w.Close()
	if err != nil {
		return err
	}
	log.Println("Quit relay")
	c.Quit()
	log.Println("E-Mail is on the way. Everything is going well.")

	return nil
}
