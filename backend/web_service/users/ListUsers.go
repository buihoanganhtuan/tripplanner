package users

import (
	"net/http"

	utils "github.com/buihoanganhtuan/tripplanner/backend/web_service/_utils"
)

type ListUserRequest struct {
	Filter string `json:"filter,omitempty"`
}

func ListUsers(w http.ResponseWriter, rq *http.Request) (error, utils.ErrorResponse) {

	return nil, utils.ErrorResponse{}
}
