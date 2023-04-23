package trips

import "net/http"

/*
***********************************************************

	Resources

***********************************************************
*/
type Trip struct {
	Id            string   `json:"id"`
	Type          string   `json:"type"`
	UserId        string   `json:"userId,omitempty"`
	Name          string   `json:"name,omitempty"`
	DateExpected  Datetime `json:"dateExpected"`
	DateCreated   Datetime `json:"dateCreated"`
	LastModified  Datetime `json:"lastModified"`
	Budget        Cost     `json:"budgetLimit"`
	PreferredMode string   `json:"preferredTransportMode"`
	PlanResult    []Edge   `json:"planResult"`
}

/************************************************************
					Requests and Responses
************************************************************/

/*
***********************************************************

	Data types

***********************************************************
*/
type Datetime struct {
	Year     int    `json:"year"`
	Month    int    `json:"month"`
	Day      int    `json:"day"`
	Hour     int    `json:"hour"`
	Min      int    `json:"min"`
	TimeZone string `json:"timezone"`
}

type Edge struct {
	PointId     string   `json:"pointId"`
	NextPointId string   `json:"nextPointId"`
	Start       Datetime `json:"start"`
	Duration    Duration `json:"duration"`
	Cost        Cost     `json:"Cost"`
	Transport   string   `json:"transportMode"`
	GeoPointId  []string `json:"geoPointId"`
}

type Duration struct {
	Hour int `json:"hour"`
	Min  int `json:"min"`
}

type Cost struct {
	Amount int    `json:"amount"`
	Unit   string `json:"unit"`
}

/*
***********************************************************

	Auxiliary types

***********************************************************
*/

type Middleware http.Handler
type MiddlewareGenerator func(http.Handler) Middleware
type MiddlewareGeneratorConfigurator func(conf map[string]interface{}) MiddlewareGenerator
type JwtMiddlewareGeneneratorConfigurator MiddlewareGeneratorConfigurator

type Status int

type StatusError struct {
	Status        Status
	Err           error
	HttpStatus    int
	ClientMessage string
}

type GraphError []string

type CycleError []GraphError
type MultiFirstError GraphError
type MultiLastError GraphError
type SimulFirstAndLastError GraphError
type UnknownNodeIdError GraphError

type PlanResult []string
type PlanResults []PlanResult

type Cycle []int
type Cycles []Cycle

type Node struct {
	Id     string
	Before []string
	After  []string
	First  bool
	Last   bool
}
