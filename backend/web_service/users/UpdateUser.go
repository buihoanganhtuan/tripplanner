package users

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	cst "github.com/buihoanganhtuan/tripplanner/backend/web_service/_constants"
	utils "github.com/buihoanganhtuan/tripplanner/backend/web_service/_utils"
)

func UpdateUser(w http.ResponseWriter, rq *http.Request) (error, utils.ErrorResponse) {
	id := mux.Vars(rq)["id"]

	var data []byte
	_, err := rq.Body.Read(data)

	if err != nil {
		return err, utils.NewUnmarshalError()
	}

	var res UserRequest
	err = json.Unmarshal(data, &res)
	if err != nil {
		return err, utils.NewUnmarshalError()
	}

	// validate fields
	// name
	if res.Name.Defined {
		rows, err := cst.Db.Query("select count(*) from ? where name = ?", cst.Ev.Var(cst.SqlUserTableVar), *res.Name.Value)
		if err != nil {
			return err, utils.NewDatabaseQueryError()
		}

		rows.Next()
		if rows.Err() != nil {
			return rows.Err(), utils.NewUnknownError()
		}
		var cnt int
		err = rows.Scan(&cnt)
		if err != nil {
			return err, utils.NewDatabaseQueryError()
		}
		if cnt == 0 {
			return errors.New("no such user"), utils.NewInvalidIdError()
		}
		_, err = cst.Db.Exec("update ? set name = ? where id = ?",
			cst.Ev.Var(cst.SqlUserTableVar),
			*res.Name.Value,
			id)
		if err != nil {
			return err, utils.NewDatabaseQueryError()
		}
	}

	rows, err := cst.Db.Query("select name, join_date from ? where id = ?", cst.Ev.Var(cst.SqlUserTableVar), id)
	if err != nil {
		return err, utils.NewDatabaseQueryError()
	}

	if !rows.Next() {
		if rows.Err() != nil {
			return rows.Err(), utils.NewDatabaseQueryError()
		}
		return errors.New("no such user"), utils.NewInvalidIdError()
	}

	var name, joinDateStr string
	rows.Scan(&name, &joinDateStr)
	t, err := time.Parse(cst.DatetimeFormat, joinDateStr)
	if err != nil {
		return err, utils.NewServerParseError()
	}

	data = nil
	data, err = json.Marshal(UserResponse{
		Id:       id,
		Name:     name,
		JoinDate: utils.JsonDateTime(t),
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
	return nil, utils.ErrorResponse{}
}
