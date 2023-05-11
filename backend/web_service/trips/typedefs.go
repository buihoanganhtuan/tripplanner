package trips

import (
	"strings"

	utils "github.com/buihoanganhtuan/tripplanner/backend/web_service/_utils"
	"github.com/buihoanganhtuan/tripplanner/backend/web_service/points"
)

/*
***********************************************************

	Resources

***********************************************************
*/
type Trip struct {
	Id            TripId              `json:"id"`
	Type          string              `json:"type"`
	UserId        string              `json:"userId,omitempty"`
	Name          string              `json:"name,omitempty"`
	DateExpected  *utils.JsonDateTime `json:"dateExpected,omitempty"`
	DateCreated   *utils.JsonDateTime `json:"dateCreated,omitempty"`
	LastModified  utils.JsonDateTime  `json:"lastModified,omitempty"`
	Budget        Cost                `json:"budgetLimit"`
	PreferredMode string              `json:"preferredTransportMode"`
	PlanResult    []Edge              `json:"planResult"`
}

/************************************************************
					Requests and Responses
************************************************************/

/*
***********************************************************

	Data types

***********************************************************
*/
type TripId string
type Datetime struct {
	Year     int    `json:"year"`
	Month    int    `json:"month"`
	Day      int    `json:"day"`
	Hour     int    `json:"hour"`
	Min      int    `json:"min"`
	TimeZone string `json:"timezone"`
}
type Edge struct {
	PointId     points.PointId      `json:"pointId"`
	NextPointId points.PointId      `json:"nextPointId"`
	Start       utils.JsonDateTime  `json:"start"`
	Duration    Duration            `json:"duration"`
	Cost        Cost                `json:"Cost"`
	Transport   string              `json:"transportMode"`
	GeoPointId  []points.GeoPointId `json:"geoPointId"`
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

type GraphError []points.PointId
type CycleError []GraphError
type MultiFirstError GraphError
type MultiLastError GraphError
type SimulFirstAndLastError GraphError
type UnknownNodeIdError GraphError

type PointOrder []points.PointId

type Cycle []int
type Cycles []Cycle

/*
***********************************************************

	Methods

***********************************************************
*/

func (ge GraphError) Error() string {
	var pids []string
	for _, pid := range ge {
		pids = append(pids, string(pid))
	}
	return strings.Join(pids, ",")
}

func (mf MultiFirstError) Error() string {
	return GraphError(mf).Error()
}

func (ml MultiLastError) Error() string {
	return GraphError(ml).Error()
}

func (un UnknownNodeIdError) Error() string {
	return GraphError(un).Error()
}

func (ce CycleError) Error() string {
	ges := []GraphError(ce)
	var sb strings.Builder
	for _, ge := range ges {
		sb.WriteString(ge.Error())
		sb.WriteString("\\n")
	}
	return sb.String()
}

func (sm SimulFirstAndLastError) Error() string {
	return GraphError(sm).Error()
}
