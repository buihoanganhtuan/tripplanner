package users

import (
	"errors"
	"log"
	"net/http"
)

func ErrorHandler(f func(w http.ResponseWriter, rq *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, rq *http.Request) {
		err := f(w, rq)
		if err != nil {
			var se StatusError
			if errors.As(err, &se) {
				w.Write([]byte(se.ClientMessage))
				w.WriteHeader(se.HttpStatus)
				log.Printf("Error code %v: %v \\n", se.Status, se.Err)
				return
			}

			log.Printf("error %v", err)
		}
	}
}