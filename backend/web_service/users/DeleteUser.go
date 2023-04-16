package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	cst "github.com/buihoanganhtuan/tripplanner/backend/web_service/_constants"
	utils "github.com/buihoanganhtuan/tripplanner/backend/web_service/_utils"
	"github.com/buihoanganhtuan/tripplanner/backend/web_service/trips"
	"github.com/gorilla/mux"
)

func DeleteUser(w http.ResponseWriter, rq *http.Request) error {
	id := mux.Vars(rq)["id"]

	mc, err := utils.ExtractClaims(rq, cst.Pk)
	if err != nil {
		return StatusError{
			Status:        InvalidToken,
			Err:           err,
			HttpStatus:    http.StatusUnauthorized,
			ClientMessage: InvalidTokenMessge,
		}
	}

	checker := jwtChecker{mapClaims: mc}

	now := time.Now().Unix()
	checker.checkClaim("iss", cst.AuthServiceName, true)
	checker.checkClaim("sub", id, true)
	checker.checkClaim("iat", now, false)
	checker.checkClaim("exp", now, true)

	if err = checker.Err(); err != nil {
		return StatusError{
			Status:        InvalidClaim,
			Err:           err,
			HttpStatus:    http.StatusUnauthorized,
			ClientMessage: fmt.Sprintf(InvalidClaimMessage, checker.errClaim),
		}
	}

	// Recursive delete all child resources
	ctx := context.Background()
	tx, err := cst.Db.BeginTx(ctx, nil)
	if err != nil {
		return StatusError{
			Status:        DatabaseTransactionError,
			Err:           err,
			HttpStatus:    http.StatusInternalServerError,
			ClientMessage: DatabaseTransactionErrorMessage,
		}
	}

	defer tx.Rollback()
	err = tx.QueryRow("select count(?) from ? where id = ?", cst.Ev.Var(cst.SqlUserTableVar), id).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return StatusError{
				Status:        NoSuchUser,
				HttpStatus:    http.StatusBadRequest,
				ClientMessage: NoSuchUserMessage,
			}
		}
		return StatusError{
			Status:        DatabaseTransactionError,
			Err:           err,
			HttpStatus:    http.StatusInternalServerError,
			ClientMessage: DatabaseTransactionErrorMessage,
		}
	}

	if err = ExecuteDeleteTransaction(tx, id); err != nil {
		return StatusError{
			Status:        DatabaseTransactionError,
			Err:           err,
			HttpStatus:    http.StatusInternalServerError,
			ClientMessage: DatabaseTransactionErrorMessage,
		}
	}

	if err = tx.Commit(); err != nil {
		return StatusError{
			Status:        DatabaseTransactionError,
			Err:           err,
			HttpStatus:    http.StatusInternalServerError,
			ClientMessage: DatabaseTransactionErrorMessage,
		}
	}

	return nil
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
