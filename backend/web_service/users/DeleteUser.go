package users

import (
	"context"
	"fmt"
	"net/http"
	"time"

	constants "github.com/buihoanganhtuan/tripplanner/backend/web_service/_constants"
	utils "github.com/buihoanganhtuan/tripplanner/backend/web_service/_utils"
	"github.com/buihoanganhtuan/tripplanner/backend/web_service/trips"
	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
)

func DeleteUser(w http.ResponseWriter, rq *http.Request) (int, string, error) {
	id := mux.Vars(rq)["id"]

	token, err := utils.ValidateAccessToken(rq, constants.PublicKey)
	if err != nil {
		return http.StatusBadRequest, "invalid access token", fmt.Errorf("invalid access token: %v", err)
	}

	mc, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return http.StatusBadRequest, "invalid access token", fmt.Errorf("invalid access token: %v", err)
	}

	checker := jwtChecker{
		mapClaims: mc,
	}

	now := time.Now().Unix()
	checker.checkClaim("iss", "auth_service", true)
	checker.checkClaim("sub", id, true)
	checker.checkClaim("iat", now, false)
	checker.checkClaim("exp", now, true)

	if err != nil {
		return http.StatusBadRequest, "invalid access token claims", fmt.Errorf("invalid access token claims %v", err)
	}

	// Recursive delete all child resources
	

	return 0, "", nil
}

func NewDeleteTransaction(resourceId string) (string, error) {
	// if a delete transaction for this resource is already ongoing, throw an error
	var ctx = context.Background()
	_, err := constants.KvStoreDelete.Get(ctx, resourceId).Result()
	if err != redis.Nil {
		return "", fmt.Errorf("another delete transaction is already ongoing for %v", resourceId)
	}

	// otherwise, create a transaction to temporarily backup the whole tree
	transactionId := resourceId + "@" + utils.GetBase32RandomString(5)


	// if any backup operation fails, delete the transaction and return an error


	// success, return the transaction id

}

func 

func ExecuteDeleteTransaction(transactionId string) error {
	rows, err := constants.Database.Query("select id from ? where userId = ?", constants.EnvironmentVariable.Var(constants.SQL_TRIP_TABLE_VAR), id)
	if err != nil {
		return err
	}
	var tripIds []string
	for rows.Next() {
		var tripId string
		err = rows.Scan(&tripId)
		if err != nil {
			return err
		}
		tripIds = append(tripIds, tripId)
	}
	for _, tripId := range tripIds {
		if err = trips.TryDelete(tripId); err != nil {
			return err
		}
	}

}