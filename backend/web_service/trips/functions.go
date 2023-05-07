package trips

import (
	"fmt"
	"net/http"

	cst "github.com/buihoanganhtuan/tripplanner/backend/web_service/_constants"
)

func newInvalidIdError() cst.ErrorResponse {
	return cst.ErrorResponse{
		Code:    http.StatusBadRequest,
		Message: "invalid resource id",
	}
}

func newServerParseError() cst.ErrorResponse {
	return cst.ErrorResponse{
		Code:    http.StatusInternalServerError,
		Message: "server-side parsing error",
	}
}

func newClientParseError(field string) cst.ErrorResponse {
	return cst.ErrorResponse{
		Code:    http.StatusBadRequest,
		Message: fmt.Sprintf("fail to parse field %s", field),
	}
}

func newDatabaseQueryError() cst.ErrorResponse {
	return cst.ErrorResponse{
		Code:    http.StatusInternalServerError,
		Message: "database query error",
	}
}

func newUnknownError() cst.ErrorResponse {
	return cst.ErrorResponse{
		Code:    http.StatusInternalServerError,
		Message: "unknown server-side error",
	}
}
