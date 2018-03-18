package state

import (
	"encoding/json"
	"io/ioutil"
)

var ActiveConfig Config

type Config struct {
	ServerPort  int     `json:"serverPort"`
	Salt        string  `json:"sneakySalt"`
	MailAPIKey  string  `json:"mailAPIKey"`
	MenuChoices Menu    `json:"menuChoices"`
	Guests      []Guest `json:"guests"`
}

type Menu struct {
	Starters []string `json:"starters"`
	Mains    []string `json:"mains"`
}

type Guest struct {
	Names []string `json:"names"`
	Day   bool     `json:"day"`
	Code  string   `json:"code"`
}

func ReadConfig() {
	raw, err := ioutil.ReadFile("config.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(raw, &ActiveConfig)
	if err != nil {
		panic(err)
	}
}
