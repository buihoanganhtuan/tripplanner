package users

import (
	"fmt"
	"net/http"
	"time"

	constants "github.com/buihoanganhtuan/tripplanner/backend/web_service/_constants"
	utils "github.com/buihoanganhtuan/tripplanner/backend/web_service/_utils"
	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
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
	rows, err := constants.Database.Query("select id from ? where userId = id", constants.EnvironmentVariable.Var(constants.PQ_TRIP_TABLE_VAR))

	return 0, "", nil
}
