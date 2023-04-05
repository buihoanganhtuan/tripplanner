package main

import (
	"io"
	"log"
	"net/http"

	users "github.com/buihoanganhtuan/tripplanner/backend/web_service/users"
	mux "github.com/gorilla/mux"
)

func ErrorHandler(f func(w http.ResponseWriter, rq *http.Request) (error, string, int)) http.HandlerFunc {
	return func(w http.ResponseWriter, rq *http.Request) {
		err, msg, code := f(w, rq)
		if err != nil {
			w.WriteHeader(code)
			io.WriteString(w, msg)
			log.Println(err)
		}
	}
}

func main() {
	r := mux.NewRouter()

	// User resource type
	r.HandleFunc("/users}", ErrorHandler(users.CreateUser)).Methods("POST")
	r.HandleFunc("/{resource.id=users/.*}/", ErrorHandler(users.UpdateUser)).Methods("PATCH")
	r.HandleFunc("/{resource.id=users/.*}/", ErrorHandler(users.ReplaceUser)).Methods("PUT")
	r.HandleFunc("/{id=users/.*}/", ErrorHandler(users.GetUser)).Methods("GET")
	r.HandleFunc("/users", ErrorHandler(users.ListUsers)).Methods("GET")
	r.HandleFunc("/{id=users/.*}/", ErrorHandler(users.DeleteUser)).Methods("DELETE")

}
