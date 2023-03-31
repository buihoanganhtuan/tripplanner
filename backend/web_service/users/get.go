package users

import (
	"net/http"
)

var usersGetHandler = ErrorHandler(_usersGetHandler)
var fileServer = http.StripPrefix("/users", http.FileServer(http.Dir("./users/assets/")))

func _usersGetHandler(w http.ResponseWriter, rq *http.Request) (error, int) {
	fileServer.ServeHTTP(w, rq)
	return nil, http.StatusOK
}
