package users

import (
	cst "github.com/buihoanganhtuan/tripplanner/backend/web_service/_constants"
	jwt "github.com/golang-jwt/jwt/v4"
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
	Id       string   `json:"id,omitempty"`
	Name     string   `json:"name,omitempty"`
	JoinDate DateTime `json:"joinDate,omitempty"`
}

type UserRequest struct {
	Id   string               `json:"id"`
	Name cst.Optional[string] `json:"name"`
}

// ********************** Data types *******************************
type DateTime struct {
	Year   string `json:"year"`
	Month  string `json:"month"`
	Day    string `json:"day"`
	Hour   string `json:"hour"`
	Min    string `json:"min"`
	Sec    string `json:"sec"`
	Offset string `json:"timezone"`
}

// ********************** Auxiliary types *******************************
type Status int

type StatusError struct {
	Status        Status
	Err           error
	HttpStatus    int
	ClientMessage string
}

type jwtChecker struct {
	mapClaims jwt.MapClaims
	errClaim  string
}
