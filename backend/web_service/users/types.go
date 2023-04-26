package users

import (
	cst "github.com/buihoanganhtuan/tripplanner/backend/web_service/_constants"
)

// var fileServer = http.StripPrefix("/users", http.FileServer(http.Dir("./users/assets/")))

// ********************** Resources *********************************
type User struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// ********************** Request/Response *************************
type UserResponse struct {
	Id       string           `json:"id,omitempty"`
	Name     string           `json:"name,omitempty"`
	JoinDate cst.JsonDateTime `json:"joinDate,omitempty"`
}

type UserRequest struct {
	Id   string               `json:"id"`
	Name cst.Optional[string] `json:"name"`
}

// ********************** Data types *******************************

// ********************** Auxiliary types *******************************
