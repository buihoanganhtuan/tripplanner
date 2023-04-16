package users

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	cst "github.com/buihoanganhtuan/tripplanner/backend/web_service/_constants"
	utils "github.com/buihoanganhtuan/tripplanner/backend/web_service/_utils"
)

func UpdateUser(w http.ResponseWriter, rq *http.Request) error {
	id := mux.Vars(rq)["id"]

	if !utils.VerifyBase32String(id, IdLengthChar) {
		return StatusError{
			Status:        InvalidId,
			HttpStatus:    http.StatusBadRequest,
			ClientMessage: InvalidIdMessage,
		}
	}

	claims, err := utils.ExtractClaims(rq, cst.Pk)

	if err != nil {
		return StatusError{
			Status:        InvalidToken,
			Err:           err,
			HttpStatus:    http.StatusBadRequest,
			ClientMessage: InvalidTokenMessge,
		}
	}

	checker := jwtChecker{
		mapClaims: claims,
	}

	now := time.Now().Unix()
	checker.checkClaim("iss", cst.AuthServiceName, true)
	checker.checkClaim("sub", id, true)
	checker.checkClaim("iat", now, false)
	checker.checkClaim("exp", now, true)

	if checker.Err() != nil {
		return StatusError{
			Status:        InvalidClaim,
			Err:           checker.Err(),
			HttpStatus:    http.StatusBadRequest,
			ClientMessage: fmt.Sprintf(InvalidClaimMessage, checker.errClaim),
		}
	}

	var data []byte
	_, err = rq.Body.Read(data)

	if err != nil {
		return StatusError{
			Status:        InvalidRequestBody,
			Err:           err,
			HttpStatus:    http.StatusBadRequest,
			ClientMessage: InvalidRequestBodyMessage,
		}
	}

	var res UserRequest
	json.Unmarshal(data, &res)

	// validate fields
	// name
	if res.Name.Defined {
		if *res.Name.Value == "" {
			return StatusError{
				Status:        InvalidFieldValue,
				Err:           errors.New("empty username"),
				HttpStatus:    http.StatusBadRequest,
				ClientMessage: fmt.Sprintf(InvalidFieldValueMessage, "name"),
			}
		}

		rows, err := cst.Db.Query("select count(*) from ? where name = ?", cst.Ev.Var(cst.SqlUserTableVar), *res.Name.Value)
		if err != nil {
			return StatusError{
				Status:        DatabaseQueryError,
				Err:           err,
				HttpStatus:    http.StatusInternalServerError,
				ClientMessage: DatabaseQueryErrorMessage,
			}
		}

		rows.Next()
		if rows.Err() != nil {
			return StatusError{
				Status:        UnknownError,
				Err:           rows.Err(),
				HttpStatus:    http.StatusInternalServerError,
				ClientMessage: UnknownErrorMessage,
			}
		}
		var cnt int
		err = rows.Scan(&cnt)
		if err != nil {
			return StatusError{
				Status:        DatabaseQueryError,
				Err:           err,
				HttpStatus:    http.StatusInternalServerError,
				ClientMessage: DatabaseQueryErrorMessage,
			}
		}
		if cnt >= 1 {
			return StatusError{
				Status:        UsernameExisted,
				HttpStatus:    http.StatusConflict,
				ClientMessage: UsernameExistedMessage,
			}
		}
		_, err = cst.Db.Exec("update ? set name = ? where id = ?",
			cst.Ev.Var(cst.SqlUserTableVar),
			*res.Name.Value,
			id)
		if err != nil {
			return StatusError{
				Status:        DatabaseQueryError,
				Err:           err,
				HttpStatus:    http.StatusInternalServerError,
				ClientMessage: DatabaseQueryErrorMessage,
			}
		}
	}

	rows, err := cst.Db.Query("select name, join_date from ? where id = ?", cst.Ev.Var(cst.SqlUserTableVar), id)
	if err != nil {
		return StatusError{
			Status:        DatabaseQueryError,
			Err:           err,
			HttpStatus:    http.StatusInternalServerError,
			ClientMessage: DatabaseQueryErrorMessage,
		}
	}

	if !rows.Next() {
		if rows.Err() != nil {
			return StatusError{
				Status:        DatabaseQueryError,
				Err:           rows.Err(),
				HttpStatus:    http.StatusInternalServerError,
				ClientMessage: DatabaseQueryErrorMessage,
			}
		}
		return StatusError{
			Status:        NoSuchUser,
			HttpStatus:    http.StatusBadRequest,
			ClientMessage: NoSuchUserMessage,
		}
	}

	var name, joinDateStr string
	rows.Scan(&name, &joinDateStr)
	t, err := time.Parse(cst.DatetimeFormat, joinDateStr)
	if err != nil {
		return StatusError{
			Status:        ParseError,
			Err:           err,
			HttpStatus:    http.StatusInternalServerError,
			ClientMessage: fmt.Sprintf(ParseErrorMessage, "joinDate"),
		}
	}

	data = nil
	var _, offset = t.Zone()
	data, err = json.Marshal(UserResponse{
		Id:   id,
		Name: name,
		JoinDate: DateTime{
			Year:   strconv.Itoa(t.Year()),
			Month:  strconv.Itoa(int(t.Month())),
			Day:    strconv.Itoa(t.Day()),
			Hour:   strconv.Itoa(t.Hour()),
			Min:    strconv.Itoa(t.Minute()),
			Offset: strconv.Itoa(offset),
		},
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
	return nil
}
