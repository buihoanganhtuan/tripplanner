package users

import (
	"log"
	"net/http"
)

func ErrorHandler(f func(w http.ResponseWriter, rq *http.Request) (error, int)) http.HandlerFunc {
	return func(w http.ResponseWriter, rq *http.Request) {
		err, statusCode := f(w, rq)
		if err != nil {
			w.WriteHeader(statusCode)
			log.Println(err)
		}
	}
}
