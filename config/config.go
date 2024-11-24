package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Addr         string
	DSN          string
	MSN          string
	MailHost     string
	MailPort     int
	MailUsername string
	MailPassword string
	MailSender   string
}

func Load(path string) (Config, error) {
	cfg := Config{
		Addr:         ":2025",
		DSN:          "kudoer.db",
		MSN:          "media_store",
		MailHost:     "",
		MailPort:     25,
		MailUsername: "",
		MailPassword: "",
		MailSender:   "Kudoer <no-reply@kudoer.com>",
	}
	_, err := toml.DecodeFile(path, &cfg)
	if err != nil {
		return Config{}, fmt.Errorf("failed loading config: %v", err)
	}
	return cfg, nil
}
