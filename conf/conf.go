package conf

import (
	"log"
	"os"
	"path"
	"strings"

	"github.com/BurntSushi/toml"
)

type Relay struct {
	Mail   string
	Secret string
	Host   string
	User   string
}

type Config struct {
	DataFileName string
	Relay        *Relay
	EmailTarget  string
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

	log.Println("Configuration: ", Current.EmailTarget, Current.Relay)
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
