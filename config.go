package main

import (
	"log"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Listen       string
	Path         string
	BaseURL      string
	Recent       int
	SaveInterval Duration
	Site         *SiteConfig
}

type SiteConfig struct {
	Title          string
	Description    string
	Keywords       KeywordList
	GuessLanguages LanguageList
}

type KeywordList []string
type LanguageList []string

type HipChatConfig struct {
	Enabled   bool
	Room      string
	RoomToken string
}

type Duration struct {
	time.Duration
}

func (k KeywordList) String() string {
	return strings.Join(k, " ")
}

func (l LanguageList) String() string {
	return "\"" + strings.Join(l, "\", \"") + "\""
}

func (d *Duration) UnmarshalText(text []byte) (err error) {
	d.Duration, err = time.ParseDuration(string(text))
	return
}

func loadConfig(configFile string) (config Config) {
	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		log.Fatal("Unable to parse configuration file: ", err)
	}
	return
}
