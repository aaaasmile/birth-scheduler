package telegram

import (
	"birthsch/conf"
	"birthsch/idl"
	"bytes"
	"fmt"
	"html/template"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramSender struct {
	cfg      conf.Telegram
	simulate bool
	content  string
	debug    bool
}

func (ts *TelegramSender) FillConf(simulate, debug bool) {
	ts.simulate = simulate
	ts.cfg = *conf.Current.Telegram
	ts.debug = debug
}

func (ts *TelegramSender) BuildMsg(templFileName string, listsrc []*idl.SchedNextItem) error {
	var partPlainContent bytes.Buffer
	tmplBody := template.Must(template.New("MailBody").ParseFiles(templFileName))
	if err := tmplBody.ExecuteTemplate(&partPlainContent, "mailPlain", listsrc); err != nil {
		return err
	}
	ts.content = partPlainContent.String()
	return nil
}

func (ts *TelegramSender) Send() error {
	if !ts.cfg.SendTelegram {
		log.Println("not send telegram")
		return nil
	}
	if ts.content == "" {
		return fmt.Errorf("telegram message content is empty")
	}
	log.Println("Telegram want to send: ", ts.content)
	if ts.simulate {
		log.Println("Telegram simulation, do nothing")
		return nil
	}
	bot, err := tgbotapi.NewBotAPI(ts.cfg.APIString)
	if err != nil {
		return err
	}
	bot.Debug = ts.debug

	log.Printf("[Telegram] Authorized on account %s", bot.Self.UserName)

	chat_id := ts.cfg.ChatID
	msg := tgbotapi.NewMessage(chat_id, ts.content)
	if _, err := bot.Send(msg); err != nil {
		return err
	}
	log.Println("[Telegram] message sent")

	return nil
}
