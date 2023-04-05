package users

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	constants "github.com/buihoanganhtuan/tripplanner/backend/web_service/_constants"
	utils "github.com/buihoanganhtuan/tripplanner/backend/web_service/_utils"
	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
)

var pk *rsa.PublicKey

const PUBLIC_KEY_PATH_VAR = "PUBLIC_KEY_PATH"
const PQ_USER_TABLE_VAR = "PQ_USER_TABLE"

var fileServer = http.StripPrefix("/users", http.FileServer(http.Dir("./users/assets/")))

var env utils.EnvironmentVariableMap

func init() {
	env.Fetch(PUBLIC_KEY_PATH_VAR, PQ_USER_TABLE_VAR)
	if env.Err() != nil {
		panic(fmt.Errorf("environment variable error: %v", env.Err()))
	}

	b, err := os.ReadFile(env.Var(PUBLIC_KEY_PATH_VAR))
	if err != nil {
		panic(fmt.Errorf("cannot read public key file: %v", err))
	}

	pk, err = jwt.ParseRSAPublicKeyFromPEM(b)
	if err != nil {
		panic(fmt.Errorf("fail to parse public key from file: %v", err))
	}
}

func CreateUser(w http.ResponseWriter, rq *http.Request) (error, string, int) {

}

func UpdateUser(w http.ResponseWriter, rq *http.Request) (error, string, int) {

}

func ReplaceUser(w http.ResponseWriter, rq *http.Request) (error, string, int) {

}

func GetUser(w http.ResponseWriter, rq *http.Request) (error, string, int) {
	vars := mux.Vars(rq)
	id := vars["id"]

	rows, err := constants.Db.Query("select * from ? where id = ?", env.Var(PQ_USER_TABLE_VAR), id)
	if err != nil {
		return fmt.Errorf("database connection error: %v", err), "", http.StatusInternalServerError
	}

	// get un from request url

	un, ok := strings.CutPrefix(rq.URL.Path, "/users/")
	if !ok || strings.ContainsRune(un, '/') {
		return nil, "", http.StatusNotFound
	}

	// check if there is any access token
	auth := rq.Header.Get("Authorization")
	if auth == "" {
		return fmt.Errorf("no access token"), "", http.StatusUnauthorized
	}

	// check access token integrity. Note that we don't support BasicAuth
	ts, ok := strings.CutPrefix(rq.Header.Get("Authorization"), "Bearer ")
	if !ok {
		return errors.New("invalid authorization header"), http.StatusUnauthorized
	}

	token, err := jwt.Parse(ts, func(token *jwt.Token) (interface{}, error) {
		return pk, nil
	}, jwt.WithValidMethods([]string{"RSA"}))

	if err != nil || !token.Valid {
		return errors.New("invalid token"), http.StatusBadRequest
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	checker := jwtChecker{
		mapClaims: claims,
	}

	now := time.Now().Unix()
	checker.checkClaim("iss", "auth_service", true)
	checker.checkClaim("sub", un, true)
	checker.checkClaim("iat", now, false)
	checker.checkClaim("exp", now, true)

	if checker.Err() != nil {
		return fmt.Errorf("error validating JWT claims: %v", err), http.StatusUnauthorized
	}
}

func ListUsers(w http.ResponseWriter, rq *http.Request) (error, string, int) {

}

func DeleteUser(w http.ResponseWriter, rq *http.Request) (error, string, int) {

}

func _usersGetHandler(w http.ResponseWriter, rq *http.Request) (error, int) {
	fileServer.ServeHTTP(w, rq)
	return nil, http.StatusOK
}

// ********************** Auxiliary types *******************************
type jwtChecker struct {
	mapClaims jwt.MapClaims
	errClaim  string
}

func (c *jwtChecker) checkClaim(name string, val interface{}, req bool) {
	if c.errClaim != "" {
		return
	}
	switch name {
	case "iss":
		sval, ok := val.(string)
		if !ok || !c.mapClaims.VerifyIssuer(sval, req) {
			c.errClaim = name
		}
	case "sub":
		sval, _ := val.(string)
		mval, _ := c.mapClaims[name].(string)
		if sval == "" || mval == "" {
			c.errClaim = name
		}
	case "iat":
		ival, ok := val.(int64)
		if !ok || !c.mapClaims.VerifyIssuedAt(ival, req) {
			c.errClaim = name
		}
	case "exp":
		ival, ok := val.(int64)
		if !ok || !c.mapClaims.VerifyExpiresAt(ival, req) {
			c.errClaim = name
		}
	default:
		v, present := c.mapClaims["name"]
		if present && v != val || !present && req {
			c.errClaim = name
		}
	}
}

func (c *jwtChecker) Err() error {
	if c.errClaim != "" {
		return fmt.Errorf("error for claim %s", c.errClaim)
	}
	return nil
}
