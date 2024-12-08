package conf

import (
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	DataFileName string
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
	return Current, nil
}
