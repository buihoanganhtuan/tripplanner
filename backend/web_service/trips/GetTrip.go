package trips

import (
	"net/http"

	cst "github.com/buihoanganhtuan/tripplanner/backend/web_service/_constants"
	utils "github.com/buihoanganhtuan/tripplanner/backend/web_service/_utils"
)

func GetAnonymousTrip(w http.ResponseWriter, rq *http.Request) (int, string, error) {

	return nil, "", 0
}

func GetRegisteredTrip(w http.ResponseWriter, rq *http.Request) (int, string, error) {
	cl, err := utils.ExtractClaims(rq, cst.Pk)

}
