package users

import (
	"net/http"
)

type ListUserRequest struct {
	Filter string `json:"filter,omitempty"`
}

func ListUsers(w http.ResponseWriter, rq *http.Request) (int, string, error) {

	return 0, "", nil
}
