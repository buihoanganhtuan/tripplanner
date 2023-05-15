package main

import (
	"github.com/buihoanganhtuan/tripplanner/backend/web_service/api/rest"
	"github.com/buihoanganhtuan/tripplanner/backend/web_service/users"
	mux "github.com/gorilla/mux"
)

func main() {
	api := rest.Rest{}

	r := mux.NewRouter()

	r.HandleFunc("/users", newValidatorMiddleware(nil)(CreateUser)).Methods("POST")
	r.HandleFunc("/{resource.id=users/.+}/", newValidatorMiddleware(nil)(UpdateUser)).Methods("PATCH")
	r.HandleFunc("/{resource.id=users/.+}/", newValidatorMiddleware(nil)(users.ReplaceUser)).Methods("PUT")
	r.HandleFunc("/{id=users/.*}/", newValidatorMiddleware(nil)(api.GetUser)).Methods("GET")
	r.HandleFunc("/users{query=\\?.+}", newValidatorMiddleware(nil)(users.ListUsers)).Methods("GET")
	r.HandleFunc("/{id=users/.+}/", newValidatorMiddleware(nil)(users.DeleteUser)).Methods("DELETE")

}
