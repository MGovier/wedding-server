package main

import (
	"net/http"
	"log"
	"github.com/MGovier/wedding-server/routes"
	"github.com/MGovier/wedding-server/state"
)

func main() {
	state.ReadConfig()
	mux := http.NewServeMux()
	mux.HandleFunc("/auth", routes.HandleAuth)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatalf("Wedding API could not start: %v", err)
	}
}
