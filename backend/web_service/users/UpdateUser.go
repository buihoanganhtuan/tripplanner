package users

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	constants "github.com/buihoanganhtuan/tripplanner/backend/web_service/_constants"
	utils "github.com/buihoanganhtuan/tripplanner/backend/web_service/_utils"
	jwt "github.com/golang-jwt/jwt/v4"
)

func UpdateUser(w http.ResponseWriter, rq *http.Request) (int, string, error) {
	id := mux.Vars(rq)["id"]

	token, err := utils.ValidateAccessToken(rq, constants.PublicKey)

	if err != nil {
		return http.StatusUnauthorized, "invalid access token", fmt.Errorf("invalid access token: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return http.StatusBadRequest, "", fmt.Errorf("error casting token.Claims to jwt.MapClaims")
	}

	checker := jwtChecker{
		mapClaims: claims,
	}

	now := time.Now().Unix()
	checker.checkClaim("iss", "auth_service", true)
	checker.checkClaim("sub", id, true)
	checker.checkClaim("iat", now, false)
	checker.checkClaim("exp", now, true)

	if checker.Err() != nil {
		return http.StatusUnauthorized, "invalid JWT claim", fmt.Errorf("error validating JWT claims: %v", err)
	}

	var data []byte
	_, err = rq.Body.Read(data)

	if err != nil {
		return http.StatusBadRequest, "error getting raw request body", fmt.Errorf("error getting raw request body: %v", err)
	}

	var res constants.UserRequest
	json.Unmarshal(data, &res)

	// validate fields
	// name
	if res.Name.Defined {
		if *res.Name.Value == "" {
			return http.StatusBadRequest, "name field, if defined, cannot be an empty string", errors.New("empty name change")
		}

		rows, err := constants.Database.Query("select count(*) from ? where name = ?",
			constants.EnvironmentVariable.Var(constants.SQL_USER_TABLE_VAR),
			*res.Name.Value)
		if err != nil {
			return http.StatusInternalServerError, "cannot query database", fmt.Errorf("cannot query database for username change check: %v", err)
		}
		rows.Next()
		if rows.Err() != nil {
			return http.StatusInternalServerError, "database error", fmt.Errorf("something is wrong with DB: %v", rows.Err())
		}
		var cnt int
		rows.Scan(&cnt)
		if cnt >= 1 {
			return http.StatusBadRequest, "username already exist", fmt.Errorf("duplicate username change %s", *res.Name.Value)
		}
		_, err = constants.Database.Exec("update ? set name = ? where id = ?",
			constants.EnvironmentVariable.Var(constants.SQL_USER_TABLE_VAR),
			*res.Name.Value,
			id)
		if err != nil {
			return http.StatusInternalServerError, "fail to update username", fmt.Errorf("fail to update username: %s", err)
		}
	}

	rows, err := constants.Database.Query("select name, join_date from ? where id = ?", constants.EnvironmentVariable.Var(constants.SQL_USER_TABLE_VAR), id)
	if err != nil {
		return http.StatusInternalServerError, "fail to retrieve updated user data", fmt.Errorf("fail to retrieve updated user data: %v", err)
	}

	var name, joinDateStr string
	if !rows.Next() {
		if rows.Err() != nil {
			return http.StatusInternalServerError, "fail to retrieve updated user data", fmt.Errorf("fail to retrieve updated user data: %v", err)
		}
		return http.StatusBadRequest, "trying to update a non-existing user", fmt.Errorf("trying to update a non-existing user: %s", id)
	}
	rows.Scan(&name, &joinDateStr)
	t, err := time.Parse(constants.DATE_TIME_FORMAT, joinDateStr)
	if err != nil {
		return http.StatusInternalServerError, "fail to parse user join date", fmt.Errorf("fail to parse user join date: %v", err)
	}

	data = nil
	var _, offset = t.Zone()
	data, err = json.Marshal(constants.UserResponse{
		Id:   id,
		Name: name,
		JoinDate: constants.DateTime{
			Year:   strconv.Itoa(t.Year()),
			Month:  strconv.Itoa(int(t.Month())),
			Day:    strconv.Itoa(t.Day()),
			Hour:   strconv.Itoa(t.Hour()),
			Min:    strconv.Itoa(t.Minute()),
			Offset: strconv.Itoa(offset),
		},
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
	return 0, "", nil
}
