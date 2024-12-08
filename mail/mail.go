package mail

import (
	"birthsch/conf"
	"birthsch/idl"
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/smtp"
)

type MailSender struct {
	relay    conf.Relay
	simulate bool
	emailTo  string
	message  bytes.Buffer
}

func (ms *MailSender) FillConf(simulate bool) {
	ms.relay = *conf.Current.Relay
	ms.emailTo = conf.Current.EmailTarget
	ms.simulate = simulate
}

func (ms *MailSender) BuildEmailMsg(templFileName string, listsrc []*idl.SchedNextItem) error {
	//bound1 := randomBoundary()
	ms.message = bytes.Buffer{}
	bound2 := randomBoundary()

	var partHTMLCont, partSubj, partPlainContent bytes.Buffer
	tmplBodyMail := template.Must(template.New("MailBody").ParseFiles(templFileName))
	if err := tmplBodyMail.ExecuteTemplate(&partHTMLCont, "mailbody", listsrc); err != nil {
		return err
	}
	if err := tmplBodyMail.ExecuteTemplate(&partSubj, "mailSubj", listsrc); err != nil {
		return err
	}

	if err := tmplBodyMail.ExecuteTemplate(&partPlainContent, "mailPlain", listsrc); err != nil {
		return err
	}

	msg := &ms.message
	msg.Write([]byte("MIME-version: 1.0;\r\n"))
	partSubj.WriteTo(msg)
	if ms.relay.Mail != "" {
		msg.Write([]byte("From: " + ms.relay.Mail + "\r\n"))
	}
	msg.Write([]byte("To: " + ms.emailTo + "\r\n"))
	msg.Write([]byte("Content-Type:  multipart/alternative; boundary=" + `"` + bound2 + `"` + "\r\n"))
	msg.Write([]byte("\r\n"))

	// plain section
	msg.Write([]byte("--" + bound2 + "\r\n"))
	msg.Write([]byte("Content-Type: text/plain; charset=\"UTF-8\"\r\n"))
	partPlainContent.WriteTo(msg)
	msg.Write([]byte("\r\n"))

	// html section
	msg.Write([]byte("--" + bound2 + "\r\n"))
	msg.Write([]byte("Content-Type: text/html; charset=\"UTF-8\"\r\n"))
	msg.Write([]byte("Content-Transfer-Encoding: base64\r\n"))
	msg.Write([]byte("\r\n"))
	partHTMLCont64 := formatRFCRawWithEnc64(partHTMLCont.Bytes())
	partHTMLCont64.WriteTo(msg)
	msg.Write([]byte("\r\n"))
	msg.Write([]byte("--" + bound2 + "--" + "\r\n"))

	if ms.simulate {
		ss := msg.String()
		maxchar := 2000
		if len(ss) > maxchar {
			ss = ss[0:maxchar]
		}
		fmt.Printf("Message is: \n%s\n", ss)
	}
	return nil
}

func (ms *MailSender) SendEmailViaRelay() error {
	log.Println("Send email using relay host")

	if ms.simulate {
		log.Println("This is a simulation, e-mail is not sent")
		return nil
	}
	len_msg := len(ms.message.Bytes())
	if len_msg == 0 {
		return fmt.Errorf("message mail is empty")
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

	log.Println("send From", ms.relay.Mail)
	if err = c.Mail(ms.relay.Mail); err != nil {
		return err
	}
	log.Println("send To", ms.emailTo)
	if err = c.Rcpt(ms.emailTo); err != nil {
		return err
	}

	w, err := c.Data()
	if err != nil {
		return err
	}
	log.Println("Send the message to the relay (len)", len_msg, ms.message.String())
	nbt, err := w.Write(ms.message.Bytes())
	if err != nil {
		return err
	}
	log.Println("Close relay after ", nbt)
	err = w.Close()
	if err != nil {
		log.Println("[SendEmailViaRelay] error ", err)
		return err
	}
	log.Println("Quit relay")
	c.Quit()
	log.Println("E-Mail is on the way. Everything is going well.")

	return nil
}

func randomBoundary() string {
	var buf [30]byte
	_, err := io.ReadFull(rand.Reader, buf[:])
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", buf[:])
}

func formatRFCRawWithEnc64(raw []byte) *bytes.Buffer {
	//  RFC 2045 formatting to 76 col
	maxLineLen := 76
	p := base64Encode(raw)
	w := &bytes.Buffer{}
	n := 0
	lineLen := 0
	for len(p)+lineLen > maxLineLen {
		w.Write(p[:maxLineLen-lineLen])
		w.Write([]byte("\r\n"))
		p = p[maxLineLen-lineLen:]
		n += maxLineLen - lineLen
		lineLen = 0
	}
	w.Write(p)
	log.Println("Buffer size: ", n+len(p))

	return w
}

func base64Encode(message []byte) []byte {
	b := make([]byte, base64.StdEncoding.EncodedLen(len(message)))
	base64.StdEncoding.Encode(b, message)
	return b
}
