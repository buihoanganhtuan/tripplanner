package postgres

import (
	"fmt"
	"net/http"
)

// This subpackage provide an implementation of the technology-agnostic domain types defined in `planner` package

// For consistency, we should select a convention for the response when errors occur and stick with it.
// The convention chosen here is: https://google.github.io/styleguide/jsoncstyleguide.xml#Reserved_Property_Names_in_the_error_object
type ErrorResponse struct {
	Code    int               `json:"code"`
	Message string            `json:"message,omitempty"`
	Errors  []ErrorDescriptor `json:"errors,omitempty"`
}
type ErrorDescriptor struct {
	Domain       string `json:"domain,omitempty"`
	Reason       string `json:"reason,omitempty"`
	Message      string `json:"message,omitempty"`
	Location     string `json:"location,omitempty"`
	LocationType string `json:"locationType,omitempty"`
}

func NewInvalidIdError() ErrorResponse {
	return ErrorResponse{
		Code:    http.StatusBadRequest,
		Message: "invalid resource id",
	}
}

func NewServerParseError() ErrorResponse {
	return ErrorResponse{
		Code:    http.StatusInternalServerError,
		Message: "server-side parsing error",
	}
}

func NewClientParseError(field string) ErrorResponse {
	return ErrorResponse{
		Code:    http.StatusBadRequest,
		Message: fmt.Sprintf("fail to parse field %s", field),
	}
}

func NewDatabaseQueryError() ErrorResponse {
	return ErrorResponse{
		Code:    http.StatusInternalServerError,
		Message: "database query error",
	}
}

func NewDatabaseTransactionError() ErrorResponse {
	return ErrorResponse{
		Code:    http.StatusInternalServerError,
		Message: "database transaction error",
	}
}

func NewUnknownError() ErrorResponse {
	return ErrorResponse{
		Code:    http.StatusInternalServerError,
		Message: "unknown server-side error",
	}
}

func NewMarshalError() ErrorResponse {
	return ErrorResponse{
		Code:    http.StatusInternalServerError,
		Message: "fail to marshal response body",
	}
}

func NewUnmarshalError() ErrorResponse {
	return ErrorResponse{
		Code:    http.StatusBadRequest,
		Message: "fail to unmarshal request body",
	}
}
