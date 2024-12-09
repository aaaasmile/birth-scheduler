package conf

import (
	"log"
	"os"
	"path"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	DataFileName  string
	Relay         *Relay
	Telegram      *Telegram
	SimulateAlarm bool
	Debug         bool
}

type Telegram struct {
	SendTelegram bool
	ChatID       int64
	APIString    string
}

type Relay struct {
	SendMail    bool
	MailFrom    string
	Secret      string
	Host        string
	User        string
	EmailTarget string
}

var Current = &Config{}

func ReadConfig(configfile string) (*Config, error) {
	_, err := os.Stat(configfile)
	if err != nil {
		return nil, err
	}
	if _, err := toml.DecodeFile(configfile, &Current); err != nil {
		return nil, err
	}
	if err := readCustomOverrideConfig(Current, configfile); err != nil {
		return nil, err
	}

	log.Println("Configuration: ", Current.Relay.MailFrom, Current.Relay.Host, Current.Relay.MailFrom, Current.Telegram.SendTelegram)
	return Current, nil
}

func readCustomOverrideConfig(Current *Config, configfile string) error {
	base := path.Base(configfile)
	dd := path.Dir(configfile)
	ext := path.Ext(configfile)
	cf := strings.Replace(base, ext, "_custom.toml", 1)
	cf_ful := path.Join(dd, cf)
	log.Println("Check for custom config ", cf_ful)
	if _, err := os.Stat(cf_ful); err != nil {
		log.Println("No custom config file found")
		return nil
	}
	log.Println("Custom config file found", cf_ful)
	_, err := toml.DecodeFile(cf_ful, Current)
	return err
}
