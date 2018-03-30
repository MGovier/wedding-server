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
	"time"
)

func RateLimitFunc(next func(w http.ResponseWriter, r *http.Request)) http.Handler {
	lmt := tollbooth.NewLimiter(0.1, &limiter.ExpirableOptions{DefaultExpirationTTL: 5 * time.Minute})
	lmt.SetBurst(5)
	lmt.SetIPLookups([]string{"X-Forwarded-For", "X-Real-IP", "RemoteAddr"})
	return tollbooth.LimitFuncHandler(lmt, next)
}

func HandleAuth(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		handleAuthPost(w, r)
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
		if guest.Code == c.Code {
			hasher := sha256.New()
			hasher.Write([]byte(guest.Code + state.ActiveConfig.Salt))
			hash := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
			cookie := &http.Cookie{
				Name:     "BM_AuthCookie",
				Value:    hash,
				Secure:   false,
				HttpOnly: true,
				MaxAge:   31536000,
			}
			http.SetCookie(w, cookie)
			w.Header().Set("Content-Type", "application/json")
			jsn, _ := json.Marshal(types.AuthResponse{
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
