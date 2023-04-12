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
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
)

func DeleteUser(w http.ResponseWriter, rq *http.Request) (int, string, error) {
	id := mux.Vars(rq)["id"]

	mc, err := utils.ExtractClaims(rq, cst.Pk)
	if err != nil {
		return http.StatusBadRequest, "invalid access token", fmt.Errorf("invalid access token: %v", err)
	}

	checker := jwtChecker{mapClaims: mc}

	now := time.Now().Unix()
	checker.checkClaim("iss", "AuthService", true)
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
	tid, err := cst.Kvs.HGet(ctx, "resources", resourceId+".delete").Result()
	t := cst.Kvs.Time(ctx).Val()
	if err != redis.Nil {
		expStr, err := cst.Kvs.HGet(ctx, "transactions", fmt.Sprintf("%s.expire", tid)).Result()
		if err != nil {
			return "", fmt.Errorf("cannot get expiration date for ongoing transaction %s", tid)
		}
		var exp time.Time
		exp, err = time.Parse(cst.DatetimeFormat, expStr)
		if err != nil {
			return "", fmt.Errorf("cannot parse expiration date %s for transaction %s", expStr, tid)
		}
		if t.Before(exp) {
			return "", fmt.Errorf("another delete transaction is already ongoing for %v", resourceId)
		}
	}

	// otherwise, create a transaction to temporarily backup the whole tree
	transactionId := utils.GetBase32RandomString(20)

	cst.Kvs.HSet(ctx, "transaction",
		transactionId, "",
		transactionId+".type", "delete",
		transactionId+".data", resourceId+";",
		transactionId+".expire", t.Add(time.Hour*time.Duration(1)).Format(cst.DatetimeFormat))

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
		err = trips.PrepareDeleteTransaction(transactionId, tid, ctx)
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
				err = trips.UnprepareDeleteTransaction(transactionId, tid, ctx)
				if err == nil {
					break
				}
			}
		}

		cst.Kvs.HDel(ctx, "transactions",
			transactionId,
			transactionId+".type",
			transactionId+".data",
			transactionId+".expire").Result()

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
	cst.Kvs.HSet(ctx, "backup", resourceId, string(b))
	return nil
}

func UnPrepareDeleteTransaction(transactionId, resourceId string, ctx context.Context) error {
	var err error
	_, err = cst.Kvs.HDel(ctx, "resources", resourceId+".delete").Result()
	return err
}

func ExecuteDeleteTransaction(transactionId, resourceId string, ctx context.Context) error {
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
		if err = trips.ExecuteDeleteTransaction(transactionId, tid, ctx); err != nil {
			return err
		}
	}
	return nil
}

func RollbackDeleteTransaction(transactionId, resourceId string, ctx context.Context) error {
	// Recover data from redis db
	v, err := cst.Kvs.HGet(ctx, "backup", resourceId).Result()
	if err != nil {
		return err
	}
	var u User
	err = json.Unmarshal([]byte(v), &u)
	if err != nil {
		return err
	}
	_, err = cst.Db.Exec(fmt.Sprintf("insert into %s (id, name, email, password) values (?, ?, ?, ?) on conflict (id) update set (id, name, email, password) = (excluded.id, excluded.name, excluded.email, excluded.password)", u.Id, u.Name, u.Email, u.Password))
	if err != nil {
		return err
	}
	return nil
}
