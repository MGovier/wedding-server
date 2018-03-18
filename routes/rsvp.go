package routes

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/MGovier/wedding-server/state"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"net/http"
)

func HandleRSVP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		handleRSVPPost(w, r)
	case "GET":
		handleRSVPGet(w, r)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
}

type rsvpPost struct {
	RSVP    bool         `json:"rsvp"`
	Message string       `json:"message"`
	Menu    []menuChoice `json:"menu"`
	Email   string       `json:"email"`
}

type menuChoice struct {
	Starter string `json:"starter"`
	Main    string `json:"main"`
	Name    string `json:"name"`
}

func handleRSVPPost(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("BM_AuthCookie")
	if err != nil {
		http.Error(w, "missing identification token", http.StatusUnauthorized)
		return
	}
	guest, err := VerifyToken(cookie.Value)
	if err != nil {
		http.Error(w, "identification token not recognised", http.StatusUnauthorized)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var p rsvpPost
	err = decoder.Decode(&p)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusBadRequest)
		return
	}
	err = validateRSVP(guest, p)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusBadRequest)
		return
	}
	// Record in file
	if p.Email != "" {
		err = sendEmail(guest, p)
		if err != nil {
			http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
			return
		}
	}
}

func handleRSVPGet(w http.ResponseWriter, r *http.Request) {

}

func validateRSVP(guest state.Guest, post rsvpPost) error {
	if !guest.Day {
		return nil
	}
	if post.RSVP == false {
		return nil
	}
	if len(post.Menu) != len(guest.Names) {
		return errors.New("menu array length unexpected")
	}
	if len(guest.Names) == 2 && post.Menu[0].Name == post.Menu[1].Name {
		return errors.New("duplicate menu names")
	}
	for _, choice := range post.Menu {
		validName := false
		for _, name := range guest.Names {
			if choice.Name == name {
				validName = true
			}
		}
		if !validName {
			return errors.New("unrecognised name")
		}
		if !isOnTheStarterMenu(choice.Starter) {
			return errors.New("unrecognised starter choice")
		}
		if !isOnTheMainMenu(choice.Main) {
			return errors.New("unrecognised main choice")
		}
	}
	return nil
}

func isOnTheStarterMenu(choice string) bool {
	for _, starter := range state.ActiveConfig.MenuChoices.Starters {
		if choice == starter {
			return true
		}
	}
	return false
}

func isOnTheMainMenu(choice string) bool {
	for _, main := range state.ActiveConfig.MenuChoices.Mains {
		if choice == main {
			return true
		}
	}
	return false
}

func sendEmail(guest state.Guest, post rsvpPost) error {
	m := mail.NewV3Mail()
	e := mail.NewEmail("Birgit and Merlin", "us@birgitandmerlin.com")
	m.SetFrom(e)

	p := mail.NewPersonalization()
	names := formatNames(guest.Names)
	p.SetSubstitution("{{name}}", names)
	tos := []*mail.Email{mail.NewEmail(names, post.Email)}
	p.AddTos(tos...)
	bcc := []*mail.Email{mail.NewEmail("Birgit and Merlin", "us@birgitandmerlin.com")}
	p.AddBCCs(bcc...)

	// Denied
	if post.RSVP == false {
		m.SetTemplateID("b2c0ffb2-38cb-42f3-a2fe-6b164ee1e9df")
	}

	// Single day guest confirmed
	if post.RSVP == true && guest.Day && len(guest.Names) == 1 {
		m.SetTemplateID("99dd386d-c213-46e8-a744-a2fac90a4450")
		p.SetSubstitution("{{guest1_starter}}", post.Menu[0].Starter)
		p.SetSubstitution("{{guest1_main}}", post.Menu[0].Main)
	}

	// Double day guest confirmed
	if post.RSVP == true && guest.Day && len(guest.Names) == 2 {
		m.SetTemplateID("7488fb6f-70a1-4523-a7ff-378bf0f3e5ab")
		p.SetSubstitution("{{guest1}}", post.Menu[0].Name)
		p.SetSubstitution("{{guest1_starter}}", post.Menu[0].Starter)
		p.SetSubstitution("{{guest1_main}}", post.Menu[0].Main)
		p.SetSubstitution("{{guest2}}", post.Menu[1].Name)
		p.SetSubstitution("{{guest2_starter}}", post.Menu[1].Starter)
		p.SetSubstitution("{{guest2_main}}", post.Menu[1].Main)
	}

	// Evening guest(s) confirmed
	if post.RSVP && !guest.Day {
		m.SetTemplateID("193d6ea0-c6e0-456a-8089-b418faacb351")
	}

	m.AddPersonalizations(p)

	request := sendgrid.GetRequest(state.ActiveConfig.MailAPIKey, "/v3/mail/send", "https://api.sendgrid.com")
	request.Method = "POST"
	request.Body = mail.GetRequestBody(m)
	response, err := sendgrid.API(request)
	if err != nil {
		return err
	} else {
		fmt.Printf("Email sent: %v\n", response.StatusCode)
	}
	return nil
}

func formatNames(names []string) string {
	switch len(names) {
	case 1:
		return names[0]
	case 2:
		return names[0] + " and " + names[1]
	case 3:
		return names[0] + ", " + names[1] + " and " + names[2]
	}
	return ""
}
