package app

import (
	"fmt"
	"net/http"

	"github.com/buihoanganhtuan/tripplanner/backend/web_service/domain"
	"github.com/buihoanganhtuan/tripplanner/backend/web_service/encoding/json"
)

// App sees external dependencies as interfaces, not implementations
type Database interface {
	GetUser(id domain.UserId) (User, error)
}

type Network interface {
	SendUser(u string) error
	ReceiveUser() string
}

// Not all external dependencies need to be interfaced. Only dependencies that
// _may_ be changed in the future need to be. It has been deciced that our app
// will use REST, so json and http are two fixed dependencies that don't need to
// be interfaced.
var App struct {
	Db Database
}

// Application implementation of domain's types
type DateTime json.DateTime

type Address struct {
	Prefecture string `json:"prefecture"`
	City       string `json:"city"`
	District   string `json:"district"`
	LandNumber string `json:"landNumber"`
}

type KeyValuePair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Cost struct {
	Amount int    `json:"amount"`
	Unit   string `json:"unit"`
}

type Duration json.Duration

type Path struct {
	PointId     domain.PointId  `json:"pointId"`
	NextPointId domain.PointId  `json:"nextPointId"`
	Start       DateTime        `json:"start"`
	Duration    Duration        `json:"duration"`
	Transports  []TransportInfo `json:"transports"`
}

type TransportInfo struct {
	Start    DateTime `json:"start"`
	Duration Duration `json:"duration"`
	Type     string   `json:"type"`
	Info     any      `json:"info"`
}

type BusInfo struct {
	Cost      Cost   `json:"cost"`
	Operator  string `json:"operator"`
	Route     string `json:"route"`
	BusNumber string `json:"busNumber"`
}

type TrainInfo struct {
	Cost     Cost   `json:"cost"`
	Operator string `json:"operator"`
	Line     string `json:"line"`
}

type WalkInfo struct {
}

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
