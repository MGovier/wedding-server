package routes

import (
	"net/http"
	"fmt"
)

func HandleAuth(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Yay")
}

