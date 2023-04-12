package trips

import "net/http"

func DeleteTrip(w http.ResponseWriter, rq *http.Request) (error, string, int) {

	return nil, "", 0
}

func NewDeleteTransaction(resourceId string) string {

}

func PrepareDeleteTransaction(transactionId, resourceId string) error {

}

func UnprepareDeleteTransaction(transactionId, resourceId string) error {

}
