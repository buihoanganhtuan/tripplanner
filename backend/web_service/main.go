package main

import (
	mux "github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()

	// User resource type
	// r.HandleFunc("/users", users.ErrorHandler(users.CreateUser)).Methods("POST")
	// r.HandleFunc("/{resource.id=users/.+}/", users.ErrorHandler(users.UpdateUser)).Methods("PATCH")
	// r.HandleFunc("/{resource.id=users/.+}/", users.ErrorHandler(users.ReplaceUser)).Methods("PUT")
	// r.HandleFunc("/{id=users/.*}/", users.ErrorHandler(users.GetUser)).Methods("GET")
	// r.HandleFunc("/users{query=\\?.+}", users.ErrorHandler(users.ListUsers)).Methods("GET")
	// r.HandleFunc("/{id=users/.+}/", users.ErrorHandler(users.DeleteUser)).Methods("DELETE")

}
