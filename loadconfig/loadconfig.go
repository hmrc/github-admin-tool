package loadconfig

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Client struct {
		Token string `yaml:token`
		Org   string `yaml:org`
	}
}

func LoadConfig(cfg *Config) {
	f, err := os.Open("config.yml")
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		fmt.Println(err)
	}
}
