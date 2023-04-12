package trips

import (
	"fmt"
	"net/http"

	cst "github.com/buihoanganhtuan/tripplanner/backend/web_service/_constants"
	utils "github.com/buihoanganhtuan/tripplanner/backend/web_service/_utils"
	"github.com/golang-jwt/jwt/v4"
)

func GetAnonymousTrip(w http.ResponseWriter, rq *http.Request) (int, string, error) {

	return nil, "", 0
}

func GetRegisteredTrip(w http.ResponseWriter, rq *http.Request) (int, string, error) {
	token, err := utils.ExtractClaims(rq, cst.Pk)
	if err != nil {
		return http.StatusUnauthorized, "invalid access token", fmt.Errorf("invalid access token %v", err)
	}
	mc, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return http.StatusUnauthorized, "invalid access token", fmt.Errorf("token claim casting error %v", err)
	}
}
