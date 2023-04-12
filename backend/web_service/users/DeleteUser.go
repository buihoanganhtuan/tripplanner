package users

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	cst "github.com/buihoanganhtuan/tripplanner/backend/web_service/_constants"
	utils "github.com/buihoanganhtuan/tripplanner/backend/web_service/_utils"
	"github.com/buihoanganhtuan/tripplanner/backend/web_service/trips"
	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
)

func DeleteUser(w http.ResponseWriter, rq *http.Request) (int, string, error) {
	id := mux.Vars(rq)["id"]

	token, err := utils.ValidateAccessToken(rq, cst.Pk)
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
	ctx := context.Background()
	v, err := cst.Kvs.HGet(ctx, "delete.ongoing", fmt.Sprintf("%s.expire", resourceId)).Result()
	t := cst.Kvs.Time(ctx).Val()
	if err != redis.Nil {
		var exp time.Time
		exp, err = time.Parse(cst.DatetimeFormat, v)
		if err != nil {
			return "", fmt.Errorf("cannot parse datetime %v", v)
		}
		if t.Before(exp) {
			return "", fmt.Errorf("another delete transaction is already ongoing for %v", resourceId)
		}
	}

	// otherwise, create a transaction to temporarily backup the whole tree
	transactionId := utils.GetBase32RandomString(5)
	cst.Kvs.HSet(ctx, "delete.ongoing", fmt.Sprintf("%s.expire", transactionId), t.Add(time.Hour*time.Duration(1)).Format(cst.DatetimeFormat))

	var rows *sql.Rows
	rows, err = cst.Db.Query("select id from ? where userId = ?", cst.Ev.Var(cst.SqlTripTableVar), resourceId)
	if err != nil {
		return "", fmt.Errorf("error querying database %v", err)
	}

	var tids []string

	for rows.Next() {
		var tid string
		rows.Scan(&tid)
		tids = append(tids, tid)
	}

	for _, tid := range tids {
		err = trips.PrepareDeleteTransaction(transactionId, tid)
		if err != nil {
			break
		}
	}

	if err == nil {
		err = PrepareDeleteTransaction(transactionId, resourceId, ctx)
	}

	// if any backup operation fails, delete the transaction and return an error
	if err != nil {
		for r := 3; r > 0; r-- {
			err = UnPrepareDeleteTransaction(transactionId, resourceId, ctx)
			if err == nil {
				break
			}
		}

		for _, tid := range tids {
			for r := 3; r > 0; r-- {
				err = trips.UnprepareDeleteTransaction(transactionId, tid)
				if err == nil {
					break
				}
			}
		}

		for r := 3; r > 0; r-- {
			err = cst.Kvs.HDel(ctx, "delete.ongoing", fmt.Sprintf("%s.expire", transactionId)).Err()
			if err == nil {
				break
			}
		}

		return "", errors.New("cannot create new delete transaction")
	}

	// success, return the transaction id
	return transactionId, nil
}

func PrepareDeleteTransaction(transactionId, resourceId string, ctx context.Context) error {
	var rows *sql.Rows
	rows, err := cst.Db.Query("select id, name, email, password from ? where id = ?", cst.Ev.Var(cst.SqlUserTableVar), resourceId)
	if err != nil {
		return err
	}
	var id, name, email, password string
	for rows.Next() {
		rows.Scan(&id, &name, &email, &password)
	}

	var b []byte
	b, err = json.Marshal(User{
		Id:       id,
		Name:     name,
		Email:    email,
		Password: password,
	})
	if err != nil {
		return err
	}
	cst.Kvs.HSet(ctx, "delete.ongoing", fmt.Sprintf("%s.data.%s", transactionId, resourceId), string(b))
	return nil
}

func UnPrepareDeleteTransaction(transactionId, resourceId string, ctx context.Context) error {

}

func ExecuteDeleteTransaction(transactionId, resourceId string) error {
	rows, err := cst.Db.Query("select id from ? where userId = ?", cst.Ev.Var(cst.SqlTripTableVar), resourceId)
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
		if err = trips.TryDelete(tid); err != nil {
			return err
		}
	}

}
