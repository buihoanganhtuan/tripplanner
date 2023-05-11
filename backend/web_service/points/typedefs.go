package points

import (
	"strconv"

	utils "github.com/buihoanganhtuan/tripplanner/backend/web_service/_utils"
)

/*
Resource types
*/
type Point struct {
	Id         PointId    `json:"id"`
	GeoPointId GeoPointId `json:"geoPointId"`
	Name       string     `json:"-"`
	Lat        float64
	Lon        float64
	Next       []PointId
	First      bool
	Last       bool
	Duration   utils.Optional[DurationConstraint]
	Arrival    utils.Optional[ArrivalConstraint]
}

type GeoPoint struct {
	Id     GeoPointId `json:"geoPointId"`
	HashId GeoHashId  `json:"-"`
	Routes []RouteId  `json:"-"`
	Lat    float64    `json:"latitude"`
	Lon    float64    `json:"longitude"`
	Tags   Tags       `json:"tags,omitempty"`
}

/*
Data types
*/
type GeoPointId int64
type PointId string
type RouteId int64
type GeoHashId int64
type Tags map[string]string

type ArrivalConstraint struct {
	Before utils.JsonDateTime `json:"before"`
}

type DurationConstraint struct {
	Duration int    `json:"duration"`
	Unit     string `json:"unit"`
}

/*
	Request/Response types
*/

/*
	Auxiliary types
*/

/*
Methods
*/
func (gpid GeoPointId) Stringer() string {
	return strconv.FormatInt(int64(gpid), 10)
}
