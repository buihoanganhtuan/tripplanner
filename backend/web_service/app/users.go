package app

import (
	"context"
	"database/sql"
	gjson "encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/buihoanganhtuan/tripplanner/backend/web_service/domain"
	"github.com/buihoanganhtuan/tripplanner/backend/web_service/encoding/json"
	"github.com/gorilla/mux"
)

// Application implementation of domain's types
type User struct {
	Id       domain.UserId `json:"id"`
	Name     string        `json:"name"`
	JoinDate json.DateTime `json:"joinDate"`
	Email    string        `json:"email"`
	Password string        `json:"password"`
}

func GetUser(id string) (domain.User, error) {
	if len(id) == 0 {
		return domain.User{}, errors.New("empty id")
	}

	u, err := App.Db.GetUser(domain.UserId(id))
	if err != nil {

	}

	var b []byte
	b, err = gjson.Marshal(u.JoinDate)
	if err != nil {

	}

	return domain.User{
		Id:       domain.UserId(id),
		Name:     u.Name,
		JoinDate: string(b),
		Email:    u.Email,
		Password: u.Password,
	}, nil
}

func UpdateUser(user domain.User) (domain.User, error) {
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
