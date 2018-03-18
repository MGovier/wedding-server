package routes

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/MGovier/wedding-server/state"
	"github.com/didip/tollbooth"
	"net/http"
)

var lmt = tollbooth.NewLimiter(1, nil)

func HandleAuth(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		handleAuthPost(w, r)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
}

type code struct {
	Code string `json:"code"`
}
type authResponse struct {
	Names []string `json:"names"`
	Day   bool     `json:"day"`
}

func handleAuthPost(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var c code
	err := decoder.Decode(&c)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusBadRequest)
		return
	}
	for _, guest := range state.ActiveConfig.Guests {
		if guest.Code == c.Code {
			hasher := sha256.New()
			hasher.Write([]byte(guest.Code + state.ActiveConfig.Salt))
			hash := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
			cookie := &http.Cookie{
				Name:     "BM_AuthCookie",
				Value:    hash,
				Secure:   true,
				HttpOnly: true,
				MaxAge:   31536000,
			}
			http.SetCookie(w, cookie)
			w.Header().Set("Content-Type", "application/json")
			jsn, _ := json.Marshal(authResponse{
				Names: guest.Names,
				Day:   guest.Day,
			})
			w.Write([]byte(jsn))
			return
		}
	}
	http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	return
}

func VerifyToken(token string) (string, error) {
	for _, guest := range state.ActiveConfig.Guests {
		hasher := sha256.New()
		hasher.Write([]byte(guest.Code))
		hash := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
		if hash == token {
			return guest.Code, nil
		}
	}
	return "", errors.New("could not find a guest for that code")
}
