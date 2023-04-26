package trips

import (
	constants "github.com/buihoanganhtuan/tripplanner/backend/web_service/_constants"
)

/*
***********************************************************

	Resources

***********************************************************
*/
type Trip struct {
	Id            string                  `json:"id"`
	Type          string                  `json:"type"`
	UserId        string                  `json:"userId,omitempty"`
	Name          string                  `json:"name,omitempty"`
	DateExpected  *constants.JsonDateTime `json:"dateExpected,omitempty"`
	DateCreated   *constants.JsonDateTime `json:"dateCreated,omitempty"`
	LastModified  constants.JsonDateTime  `json:"lastModified,omitempty"`
	Budget        Cost                    `json:"budgetLimit"`
	PreferredMode string                  `json:"preferredTransportMode"`
	PlanResult    []Edge                  `json:"planResult"`
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
	PointId     PointId                `json:"pointId"`
	NextPointId PointId                `json:"nextPointId"`
	Start       constants.JsonDateTime `json:"start"`
	Duration    Duration               `json:"duration"`
	Cost        Cost                   `json:"Cost"`
	Transport   string                 `json:"transportMode"`
	GeoPointId  []GeoPointId           `json:"geoPointId"`
}

type Duration struct {
	Hour int `json:"hour"`
	Min  int `json:"min"`
}

type Cost struct {
	Amount int    `json:"amount"`
	Unit   string `json:"unit"`
}

type PointId string
type GeoPointId string

/*
***********************************************************

	Auxiliary types

***********************************************************
*/

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

type NodeId string
type Node struct {
	Id     NodeId
	Before []NodeId
	After  []NodeId
	First  bool
	Last   bool
}
