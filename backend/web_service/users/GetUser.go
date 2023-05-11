package users

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	cst "github.com/buihoanganhtuan/tripplanner/backend/web_service/_constants"
	utils "github.com/buihoanganhtuan/tripplanner/backend/web_service/_utils"
	"github.com/gorilla/mux"
)

func GetUser(w http.ResponseWriter, rq *http.Request) (error, utils.ErrorResponse) {
	id := mux.Vars(rq)["id"]

	var uid, name, joindateStr string
	err := cst.Db.QueryRow("select id, name, join_date from ? where id = ?", cst.Ev.Var(cst.SqlUserTableVar), id).
		Scan(&uid, &name, &joindateStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return err, utils.NewInvalidIdError()
		}
		return err, utils.NewDatabaseQueryError()
	}

	joinDate, err := time.Parse(cst.DatetimeFormat, joindateStr)

	if err != nil {
		return err, utils.NewServerParseError()
	}

	resource, err := json.Marshal(UserResponse{
		Id:       uid,
		Name:     name,
		JoinDate: utils.JsonDateTime(joinDate),
	})

	if err != nil {
		return err, utils.NewMarshalError()
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(resource)

	return nil, utils.ErrorResponse{}
}
