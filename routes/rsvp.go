package routes

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/MGovier/wedding-server/state"
	"github.com/MGovier/wedding-server/types"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"net/http"
	"strings"
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
	_, err = state.GetData(guest.Code)
	if err == nil {
		// Hey, you already RSVPd!
		http.Error(w, "already received RSVP for this guest", http.StatusBadRequest)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var p types.RSVPPost
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
	p.Names = guest.Names
	p.Day = guest.Day
	err = state.RecordRSVP(guest, p)
	if err != nil {
		http.Error(w, "could not store data", http.StatusInternalServerError)
		fmt.Println(err)
	}
	if p.Email != "" {
		err = sendEmail(guest, p)
		if err != nil {
			http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
			return
		}
	}
	w.Write([]byte("OK"))
}

func handleRSVPGet(w http.ResponseWriter, r *http.Request) {
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
	data, err := state.GetData(guest.Code)
	if err != nil {
		jsn, _ := json.Marshal(types.AuthResponse{
			Names: guest.Names,
			Day:   guest.Day,
		})
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsn)
		return
	}
	res, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "could not marshal JSON RSVP data", http.StatusInternalServerError)
		return
	}
	w.Write(res)
}

func validateRSVP(guest types.Guest, post types.RSVPPost) error {
	if !guest.Day && (post.Guests[0].Main != "" || post.Guests[0].Starter != "") {
		return errors.New("day guests should not pick menu choices")
	}
	if len(post.Guests) != len(guest.Names) {
		return errors.New("guest array length unexpected")
	}
	if len(guest.Names) == 2 && post.Guests[0].Name == post.Guests[1].Name {
		return errors.New("duplicate guest names")
	}
	for _, g := range post.Guests {
		validName := false
		for _, name := range guest.Names {
			if g.Name == name {
				validName = true
			}
		}
		if g.Attending == nil {
			return errors.New("missing attendance")
		}
		if !validName {
			return errors.New("unrecognised name")
		}
		if *g.Attending && guest.Day {
			if !isOnTheStarterMenu(g.Starter) {
				return errors.New("unrecognised starter")
			}
			if !isOnTheMainMenu(g.Main) {
				return errors.New("unrecognised main")
			}
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

func sendEmail(guest types.Guest, post types.RSVPPost) error {
	m := mail.NewV3Mail()
	e := mail.NewEmail("Birgit and Merlin", "us@birgitandmerlin.com")
	m.SetFrom(e)

	p := mail.NewPersonalization()
	names := formatNames(guest.Names)
	p.SetSubstitution("{{name}}", names)
	// Can't have commas in email to field:
	tos := []*mail.Email{mail.NewEmail(strings.Replace(names, ",", "", -1), post.Email)}
	p.AddTos(tos...)
	bcc := []*mail.Email{mail.NewEmail("Birgit and Merlin", "us@birgitandmerlin.com")}
	p.AddBCCs(bcc...)

	rsvp := false
	for _, guest := range post.Guests {
		if *guest.Attending {
			rsvp = true
		}
	}

	if !rsvp {
		m.SetTemplateID("b2c0ffb2-38cb-42f3-a2fe-6b164ee1e9df")
	}

	// Single day guest confirmed
	if rsvp == true && guest.Day && len(guest.Names) == 1 {
		m.SetTemplateID("99dd386d-c213-46e8-a744-a2fac90a4450")
		p.SetSubstitution("{{guest1_starter}}", post.Guests[0].Starter)
		p.SetSubstitution("{{guest1_main}}", post.Guests[0].Main)
	}

	// Double day guest confirmed
	if rsvp == true && guest.Day && len(guest.Names) == 2 {
		m.SetTemplateID("7488fb6f-70a1-4523-a7ff-378bf0f3e5ab")
		p.SetSubstitution("{{guest1}}", post.Guests[0].Name)
		if !*post.Guests[0].Attending {
			p.SetSubstitution("{{guest1_starter}}", "Not attending")
			p.SetSubstitution("{{guest1_main}}", "No food please!")
		} else {
			p.SetSubstitution("{{guest1_starter}}", post.Guests[0].Starter)
			p.SetSubstitution("{{guest1_main}}", post.Guests[0].Main)
		}
		p.SetSubstitution("{{guest2}}", post.Guests[1].Name)
		if !*post.Guests[1].Attending {
			p.SetSubstitution("{{guest2_starter}}", "Not attending")
			p.SetSubstitution("{{guest2_main}}", "No food please!")
		} else {
			p.SetSubstitution("{{guest2_starter}}", post.Guests[1].Starter)
			p.SetSubstitution("{{guest2_main}}", post.Guests[1].Main)
		}
	}

	// Triple day guest confirmed
	if rsvp == true && guest.Day && len(guest.Names) == 3 {
		m.SetTemplateID("eee2b2c6-41f4-4a5c-8065-ae2cd352f029")
		p.SetSubstitution("{{guest1}}", post.Guests[0].Name)
		if !*post.Guests[0].Attending {
			p.SetSubstitution("{{guest1_starter}}", "Not attending")
			p.SetSubstitution("{{guest1_main}}", "No food please!")
		} else {
			p.SetSubstitution("{{guest1_starter}}", post.Guests[0].Starter)
			p.SetSubstitution("{{guest1_main}}", post.Guests[0].Main)
		}
		p.SetSubstitution("{{guest2}}", post.Guests[1].Name)
		if !*post.Guests[1].Attending {
			p.SetSubstitution("{{guest2_starter}}", "Not attending")
			p.SetSubstitution("{{guest2_main}}", "No food please!")
		} else {
			p.SetSubstitution("{{guest2_starter}}", post.Guests[1].Starter)
			p.SetSubstitution("{{guest2_main}}", post.Guests[1].Main)
		}
		p.SetSubstitution("{{guest3}}", post.Guests[2].Name)
		if !*post.Guests[2].Attending {
			p.SetSubstitution("{{guest3_starter}}", "Not attending")
			p.SetSubstitution("{{guest3_main}}", "No food please!")
		} else {
			p.SetSubstitution("{{guest3_starter}}", post.Guests[2].Starter)
			p.SetSubstitution("{{guest3_main}}", post.Guests[2].Main)
		}
	}

	// Evening guest(s) confirmed
	if rsvp && !guest.Day {
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
		return names[0] + ", " + names[1] + ", and " + names[2]
	}
	return ""
}
