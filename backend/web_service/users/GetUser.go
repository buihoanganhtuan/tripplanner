package users

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	constants "github.com/buihoanganhtuan/tripplanner/backend/web_service/_constants"
	"github.com/gorilla/mux"
)

func GetUser(w http.ResponseWriter, rq *http.Request) (int, string, error) {
	id := mux.Vars(rq)["id"]

	rows, err := constants.Database.Query("select id, name, join_date from ? where id = ?", constants.EnvironmentVariable.Var(constants.PQ_USER_TABLE_VAR), id)
	if err != nil {
		return http.StatusInternalServerError, "", fmt.Errorf("database connection error: %v", err)
	}

	if !rows.Next() {
		return http.StatusInternalServerError, "user not found", errors.New("valid access token but user not found in DB")
	}

	var uid, name, joinDate string
	rows.Scan(&uid, &name, &joinDate)

	t, err := time.Parse(constants.DATE_TIME_FORMAT, joinDate)

	if err != nil {
		return http.StatusInternalServerError, "error parsing join date", fmt.Errorf("error parsing join date: %v", err)
	}

	_, offset := t.Zone()

	resource, err := json.Marshal(constants.UserResponse{
		Id:   uid,
		Name: name,
		JoinDate: constants.DateTime{
			Year:   strconv.Itoa(t.Year()),
			Month:  strconv.Itoa(int(t.Month())),
			Day:    strconv.Itoa(t.Day()),
			Hour:   strconv.Itoa(t.Hour()),
			Min:    strconv.Itoa(t.Minute()),
			Offset: strconv.Itoa(offset),
		},
	})

	if err != nil {
		return http.StatusInternalServerError, "error marshalling resource", fmt.Errorf("error marshalling resource: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(resource)

	return 0, "", nil
}
