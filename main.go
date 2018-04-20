package main

import (
	"github.com/MGovier/wedding-server/routes"
	"github.com/MGovier/wedding-server/state"
	"log"
	"net/http"
)

func main() {
	state.ReadConfig()
	state.LoadData()
	mux := http.NewServeMux()

	mux.Handle("/auth", routes.RateLimitFunc(routes.HandleAuth))
	mux.HandleFunc("/rsvp", routes.HandleRSVP)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatalf("Wedding API could not start: %v", err)
	}
}
