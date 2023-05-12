package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	cst "github.com/buihoanganhtuan/tripplanner/backend/web_service/_constants"
	utils "github.com/buihoanganhtuan/tripplanner/backend/web_service/_utils"
	"github.com/buihoanganhtuan/tripplanner/backend/web_service/planner"
	"github.com/gorilla/mux"
)

func GetUser(w http.ResponseWriter, rq *http.Request) (error, planner.ErrorResponse) {
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

	return nil, planner.ErrorResponse{}
}

func UpdateUser(w http.ResponseWriter, rq *http.Request) (error, planner.ErrorResponse) {
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
		var count int
		err = rows.Scan(&count)
		if err != nil {
			return err, utils.NewDatabaseQueryError()
		}
		if count == 0 {
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
	return nil, planner.ErrorResponse{}
}

func DeleteUser(w http.ResponseWriter, rq *http.Request) (error, planner.ErrorResponse) {
	id := mux.Vars(rq)["id"]

	// Recursive delete all child resources
	ctx := context.Background()
	tx, err := cst.Db.BeginTx(ctx, nil)
	if err != nil {
		return err, utils.NewDatabaseTransactionError()
	}

	defer tx.Rollback()
	err = tx.QueryRow("select count(?) from ? where id = ?", cst.Ev.Var(cst.SqlUserTableVar), id).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return err, utils.NewInvalidIdError()
		}
		return err, utils.NewDatabaseTransactionError()
	}

	if err = ExecuteDeleteTransaction(tx, id); err != nil {
		return err, utils.NewDatabaseTransactionError()
	}

	if err = tx.Commit(); err != nil {
		return err, utils.NewDatabaseTransactionError()
	}

	return nil, planner.ErrorResponse{}
}

func ExecuteDeleteTransaction(tx *sql.Tx, resourceId string) error {
	rows, err := tx.Query("select id from ? where userId = ?", cst.Ev.Var(cst.SqlTripTableVar), resourceId)
	if err != nil {
		return err
	}

	var tids []string
	for rows.Next() {
		var tid string
		err = rows.Scan(&tid)
		if err != nil {
			return err
		}
		tids = append(tids, tid)
	}
	for _, tid := range tids {
		if err = trips.ExecuteDeleteTransaction(tx, tid); err != nil {
			return err
		}
	}
	return nil
}
