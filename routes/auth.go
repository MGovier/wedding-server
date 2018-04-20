package routes

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/MGovier/wedding-server/state"
	"github.com/MGovier/wedding-server/types"
	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"net/http"
	"strings"
	"time"
)

func RateLimitFunc(next func(w http.ResponseWriter, r *http.Request)) http.Handler {
	lmt := tollbooth.NewLimiter(0.2, &limiter.ExpirableOptions{DefaultExpirationTTL: 5 * time.Minute})
	lmt.SetBurst(5)
	lmt.SetIPLookups([]string{"X-Forwarded-For", "X-Real-IP"})
	return tollbooth.LimitFuncHandler(lmt, next)
}

func HandleAuth(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		handleAuthPost(w, r)
	case "DELETE":
		handleDelete(w, r)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
}

func handleAuthPost(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var c types.Code
	err := decoder.Decode(&c)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusBadRequest)
		return
	}
	for _, guest := range state.ActiveConfig.Guests {
		if strings.ToUpper(guest.Code) == strings.ToUpper(c.Code) {
			hasher := sha256.New()
			hasher.Write([]byte(guest.Code + state.ActiveConfig.Salt))
			hash := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
			cookie := &http.Cookie{
				Name:     "BM_AuthCookie",
				Value:    hash,
				Secure:   true,
				HttpOnly: true,
				MaxAge:   31536000,
				Path:     "/",
				Expires:  time.Now().AddDate(1, 0, 0),
				Domain:   "birgitandmerlin.com",
			}
			http.SetCookie(w, cookie)
			w.Header().Set("Content-Type", "application/json")
			data, err := state.GetData(guest.Code)
			if err != nil {
				jsn, _ := json.Marshal(types.AuthResponse{
					Names: guest.Names,
					Day:   guest.Day,
				})
				w.Write(jsn)
				return
			}
			jsn, err := json.Marshal(data)
			if err != nil {
				http.Error(w, "could not marshal JSON RSVP data", http.StatusInternalServerError)
				return
			}
			w.Write([]byte(jsn))
			return
		}
	}
	http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	return
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:     "BM_AuthCookie",
		Value:    "Deleted",
		Secure:   false,
		HttpOnly: true,
		MaxAge:   -1,
	}
	http.SetCookie(w, cookie)
	return
}

func VerifyToken(token string) (types.Guest, error) {
	for _, guest := range state.ActiveConfig.Guests {
		hasher := sha256.New()
		hasher.Write([]byte(guest.Code + state.ActiveConfig.Salt))
		hash := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
		if hash == token {
			return guest, nil
		}
	}
	return types.Guest{}, errors.New("could not find a guest for that code")
}
