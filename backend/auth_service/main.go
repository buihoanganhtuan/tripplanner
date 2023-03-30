package main

import (
	"net/http"

	"github.com/buihoanganhtuan/tripplanner/backend/auth_service/users"
)

func main() {
	http.HandleFunc("/users/", users.UsersHandler)
	http.ListenAndServe(":80", nil)
}
