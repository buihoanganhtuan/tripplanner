package main

import (
	utils "github.com/buihoanganhtuan/tripplanner/backend/web_service/_utils"
	users "github.com/buihoanganhtuan/tripplanner/backend/web_service/users"
	mux "github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()

	// User resource type
	r.HandleFunc("/users", utils.ErrorHandler(users.CreateUser)).Methods("POST")
	r.HandleFunc("/{resource.id=users/.+}/", utils.ErrorHandler(users.UpdateUser)).Methods("PATCH")
	r.HandleFunc("/{resource.id=users/.+}/", utils.ErrorHandler(users.ReplaceUser)).Methods("PUT")
	r.HandleFunc("/{id=users/.*}/", utils.ErrorHandler(users.GetUser)).Methods("GET")
	r.HandleFunc("/users{query=\\?.+}", utils.ErrorHandler(users.ListUsers)).Methods("GET")
	r.HandleFunc("/{id=users/.+}/", utils.ErrorHandler(users.DeleteUser)).Methods("DELETE")

}
