package state

import (
	"io/ioutil"
	"encoding/json"
)

var ActiveConfig Config

type Config struct {
	ServerPort int `json:"serverPort"`
	MenuChoices Menu `json:"menuChoices"`
	Guests []Guest `json:"guests"`
}

type Menu struct {
	Starters []string `json:"starters"`
	Mains []string `json:"mains"`
}

type Guest struct {
	Names []string `json:"names"`
	Code string `json:"code"`
}

func ReadConfig() {
	raw, err := ioutil.ReadFile("state.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(raw, &ActiveConfig)
	if err != nil {
		panic(err)
	}
}
