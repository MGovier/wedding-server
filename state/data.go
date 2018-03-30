package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/MGovier/wedding-server/types"
	"io/ioutil"
)

var ActiveData map[string]types.RSVPPost

func LoadData() {
	data, err := ioutil.ReadFile("data.json")
	ActiveData = make(map[string]types.RSVPPost)
	if err != nil {
		fmt.Println("Could not read data file!")
		return
	}
	var jsonArray []types.RSVPWithCode
	err = json.Unmarshal(data, &jsonArray)
	if err != nil {
		fmt.Println("Could not unmarshall data file!")
		return
	}
	for _, v := range jsonArray {
		ActiveData[v.Code] = v.RSVP
	}
	fmt.Println(ActiveData)
}

func saveData() {
	jsonArray := make([]types.RSVPWithCode, 0)
	for k, v := range ActiveData {
		jsonArray = append(jsonArray, types.RSVPWithCode{
			Code: k,
			RSVP: v,
		})
	}
	content, err := json.Marshal(jsonArray)
	if err != nil {
		fmt.Printf("could not marshal active data: %v", err)
		return
	}

	err = ioutil.WriteFile("data.json", content, 0600)
	if err != nil {
		fmt.Printf("could not write active data: %v", err)
		return
	}
}

func RecordRSVP(code string, rsvp types.RSVPPost) error {
	if code == "" {
		return errors.New("invalid code")
	}
	ActiveData[code] = rsvp
	saveData()
	return nil
}

func GetData(code string) (types.RSVPPost, error) {
	data, ok := ActiveData[code]
	if !ok {
		return types.RSVPPost{}, errors.New("guest not found in data")
	}
	return data, nil
}
