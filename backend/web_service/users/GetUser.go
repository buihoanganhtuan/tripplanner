package users

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	cst "github.com/buihoanganhtuan/tripplanner/backend/web_service/_constants"
	utils "github.com/buihoanganhtuan/tripplanner/backend/web_service/_utils"
	"github.com/gorilla/mux"
)

func GetUser(w http.ResponseWriter, rq *http.Request) error {
	id := mux.Vars(rq)["id"]

	if !utils.VerifyBase32String(id, IdLengthChar) {
		return StatusError{
			Status:        InvalidId,
			HttpStatus:    http.StatusBadRequest,
			ClientMessage: InvalidIdMessage,
		}
	}

	var uid, name, _joinDate string
	err := cst.Db.QueryRow("select id, name, join_date from ? where id = ?", cst.Ev.Var(cst.SqlUserTableVar), id).
		Scan(&uid, &name, &_joinDate)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return StatusError{
				Status:        NoSuchUser,
				HttpStatus:    http.StatusBadRequest,
				ClientMessage: NoSuchUserMessage,
			}
		}
		return StatusError{
			Status:        DatabaseQueryError,
			Err:           err,
			HttpStatus:    http.StatusInternalServerError,
			ClientMessage: DatabaseQueryErrorMessage,
		}
	}

	joinDate, err := time.Parse(cst.DatetimeFormat, _joinDate)

	if err != nil {
		return StatusError{
			Status:        ParseError,
			Err:           err,
			HttpStatus:    http.StatusInternalServerError,
			ClientMessage: fmt.Sprintf(ParseErrorMessage, "joinDate"),
		}
	}

	_, offset := joinDate.Zone()

	resource, err := json.Marshal(UserResponse{
		Id:   uid,
		Name: name,
		JoinDate: DateTime{
			Year:   strconv.Itoa(joinDate.Year()),
			Month:  strconv.Itoa(int(joinDate.Month())),
			Day:    strconv.Itoa(joinDate.Day()),
			Hour:   strconv.Itoa(joinDate.Hour()),
			Min:    strconv.Itoa(joinDate.Minute()),
			Offset: strconv.Itoa(offset),
		},
	})

	if err != nil {
		return StatusError{
			Status:        MarshallingError,
			Err:           err,
			HttpStatus:    http.StatusInternalServerError,
			ClientMessage: MarshallingErrorMessage,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(resource)

	return nil
}
