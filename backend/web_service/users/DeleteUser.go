package users

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	cst "github.com/buihoanganhtuan/tripplanner/backend/web_service/_constants"
	utils "github.com/buihoanganhtuan/tripplanner/backend/web_service/_utils"
	"github.com/buihoanganhtuan/tripplanner/backend/web_service/trips"
	"github.com/gorilla/mux"
)

func DeleteUser(w http.ResponseWriter, rq *http.Request) (error, utils.ErrorResponse) {
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

	return nil, utils.ErrorResponse{}
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
