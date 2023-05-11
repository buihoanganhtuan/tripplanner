package users

import (
	utils "github.com/buihoanganhtuan/tripplanner/backend/web_service/_utils"
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
	Id       string             `json:"id,omitempty"`
	Name     string             `json:"name,omitempty"`
	JoinDate utils.JsonDateTime `json:"joinDate,omitempty"`
}

type UserRequest struct {
	Id   string                 `json:"id"`
	Name utils.Optional[string] `json:"name"`
}

// ********************** Data types *******************************

// ********************** Auxiliary types *******************************
