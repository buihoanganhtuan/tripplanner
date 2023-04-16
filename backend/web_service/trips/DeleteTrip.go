package trips

import (
	"context"
	"database/sql"
	"net/http"
)

func DeleteTrip(w http.ResponseWriter, rq *http.Request) (error, string, int) {

	return nil, "", 0
}

func NewDeleteTransaction(resourceId string) string {

}

func PrepareDeleteTransaction(transactionId, resourceId string, ctx context.Context) error {

}

func UnprepareDeleteTransaction(transactionId, resourceId string, ctx context.Context) error {

}

func ExecuteDeleteTransaction(tx *sql.Tx, resourceId string) error {

}
