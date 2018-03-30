package state

import (
	"encoding/json"
	"github.com/MGovier/wedding-server/types"
	"io/ioutil"
)

var ActiveConfig types.Config

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
