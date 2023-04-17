package trips

import (
	"fmt"
	"net/http"

	cst "github.com/buihoanganhtuan/tripplanner/backend/web_service/_constants"
	utils "github.com/buihoanganhtuan/tripplanner/backend/web_service/_utils"
	"github.com/gorilla/mux"
)

func PlanAnonymousTrip(w http.ResponseWriter, rq *http.Request) error {
	return nil
}

func PlanRegisteredTrip(w http.ResponseWriter, rq *http.Request) error {
	id := mux.Vars(rq)["id"]

	if !utils.VerifyBase32String(id, IdLengthChar) {
		return StatusError{
			Status:        InvalidId,
			HttpStatus:    http.StatusBadRequest,
			ClientMessage: InvalidIdMessage,
		}
	}

	cm, err := utils.ExtractClaims(rq, cst.Pk)
	if err != nil {
		return StatusError{
			Status:        InvalidToken,
			Err:           err,
			HttpStatus:    http.StatusBadRequest,
			ClientMessage: InvalidTokenMessge,
		}
	}

	ok, cln := verifyJwtClaims(cm)
	if !ok {
		return StatusError{
			Status:        InvalidClaim,
			HttpStatus:    http.StatusBadRequest,
			ClientMessage: fmt.Sprintf(InvalidClaimMessage, cln),
		}
	}

}
