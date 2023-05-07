package trips

import (
	"strconv"
	"strings"

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

type Point struct {
	Id         PointId
	GeoPointId GeoPointId
	Name       string
	Lat        float64
	Lon        float64
	Next       []Point
	First      bool
	Last       bool
}

type GeoPoint struct {
	Id     GeoPointId `json:"geoPointId"`
	HashId GeoHashId  `json:"-"`
	Routes []RouteId  `json:"-"`
	Lat    float64    `json:"latitude"`
	Lon    float64    `json:"longitude"`
	Tags   Tags       `json:"tags,omitempty"`
}

type ArrivalConstraint struct {
	from constants.JsonDateTime `json:"from"`
	to   constants.JsonDateTime `json:"to"`
}

type DurationConstraint struct {
}

type PointId int64
type GeoPointId int64
type Tags map[string]string

/*
***********************************************************

	Auxiliary types

***********************************************************
*/

type GraphError []PointId
type CycleError []GraphError
type MultiFirstError GraphError
type MultiLastError GraphError
type SimulFirstAndLastError GraphError
type UnknownNodeIdError GraphError

type PointOrder []PointId

type Cycle []int
type Cycles []Cycle

type RouteId int64
type GeoHashId int64

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

func (gpid GeoPointId) Stringer() string {
	return strconv.FormatInt(int64(gpid), 10)
}
