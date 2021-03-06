package types

type RSVPWithCode struct {
	Code string   `json:"code"`
	RSVP RSVPPost `json:"rsvp"`
}

type RSVPPost struct {
	Message string        `json:"message"`
	Guests  []GuestDetail `json:"guests"`
	Email   string        `json:"email"`
	Names   []string      `json:"names,omitempty"`
	Day     bool          `json:"day"`
}

type GuestDetail struct {
	Starter   string `json:"starter"`
	Main      string `json:"main"`
	Name      string `json:"name"`
	Attending *bool  `json:"attending,omitempty"`
}

type Code struct {
	Code string `json:"code"`
}

type AuthResponse struct {
	Names []string `json:"names"`
	Day   bool     `json:"day"`
}

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
