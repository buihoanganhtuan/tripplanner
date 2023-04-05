package users

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	constants "github.com/buihoanganhtuan/tripplanner/backend/web_service/_constants"
	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
)

func GetUser(w http.ResponseWriter, rq *http.Request) (error, string, int) {
	vars := mux.Vars(rq)
	id := vars["id"]

	// check if there is any access token
	if rq.Header.Get("Authorization") == "" {
		return fmt.Errorf("no access token"), "", http.StatusUnauthorized
	}

	// check access token integrity. Note that we don't support BasicAuth
	ts, ok := strings.CutPrefix(rq.Header.Get("Authorization"), "Bearer ")
	if !ok {
		return errors.New("invalid authorization header"), "invalid authorization header", http.StatusUnauthorized
	}

	token, err := jwt.Parse(ts, func(token *jwt.Token) (interface{}, error) {
		return constants.PublicKey, nil
	}, jwt.WithValidMethods([]string{"RSA"}))

	if err != nil || !token.Valid {
		return errors.New("invalid token"), "invalid access token", http.StatusBadRequest
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	checker := jwtChecker{
		mapClaims: claims,
	}

	now := time.Now().Unix()
	checker.checkClaim("iss", "auth_service", true)
	checker.checkClaim("sub", id, true)
	checker.checkClaim("iat", now, false)
	checker.checkClaim("exp", now, true)

	if checker.Err() != nil {
		return fmt.Errorf("error validating JWT claims: %v", err), "invalid JWT claim", http.StatusUnauthorized
	}

	rows, err := constants.Database.Query("select id, name, email, join_date from ? where id = ?", env.Var(constants.PQ_USER_TABLE_VAR), id)
	if err != nil {
		return fmt.Errorf("database connection error: %v", err), "", http.StatusInternalServerError
	}

	if !rows.Next() {
		return errors.New("valid access token but user not found in DB"), "user not found", http.StatusInternalServerError
	}

	var uid, name, email, joinDate string
	rows.Scan(&uid, &name, &email, &joinDate)

	t, err := time.Parse(constants.DATE_TIME_FORMAT, joinDate)

	if err != nil {
		return fmt.Errorf("error parsing join date: %v", err), "error parsing join date", http.StatusInternalServerError
	}

	_, offset := t.Zone()
	jd := constants.DateTime{
		Year:   strconv.Itoa(t.Year()),
		Month:  strconv.Itoa(int(t.Month())),
		Day:    strconv.Itoa(t.Day()),
		Hour:   strconv.Itoa(t.Hour()),
		Min:    strconv.Itoa(t.Minute()),
		Offset: strconv.Itoa(offset),
	}

	resource, err := json.Marshal(constants.User{
		Id:       uid,
		Name:     name,
		Email:    email,
		JoinDate: jd,
	})

	if err != nil {
		return fmt.Errorf("error marshalling resource: %v", err), "error marshalling resource", http.StatusInternalServerError
	}

	w.Write(resource)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	return nil, "", 0
}
