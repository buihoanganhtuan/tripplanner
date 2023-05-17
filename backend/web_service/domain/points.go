package domain

import "errors"

const (
	GeohashLen = 41
)

type Point struct {
	Id         PointId                 `json:"id"`
	TripId     TripId                  `json:"tripId"`
	GeoPointId GeoPointId              `json:"geoPointId"`
	Arrival    *PointArrivalConstraint `json:"arrivalConstraint"`
	Duration   Duration                `json:"durationConstraint"`
	Before     PointBeforeConstraint   `json:"beforeConstraint"`
	After      PointAfterConstraint    `json:"afterConstraint"`
	First      bool                    `json:"isFirst"`
	Last       bool                    `json:"isLast"`
}

type PointId string

type PointArrivalConstraint struct {
	Before DateTime
}

type PointAfterConstraint struct {
	Points []PointId
}

type PointBeforeConstraint struct {
	Points []PointId `json:"points"`
}

type GeoPoint struct {
	Id      GeoPointId     `json:"id"`
	Lat     float64        `json:"lat"`
	Lon     float64        `json:"lon"`
	Name    *string        `json:"name,omitempty"`
	Address Address        `json:"address"`
	Tags    []KeyValuePair `json:"tags,omitempty"`
}

type GeoPointId string

type RouteId string

type Route struct {
	Id         RouteId        `json:"id"`
	GeoPoints  []GeoPoint     `json:"nodes"`
	Transports []string       `json:"transport"`
	Tags       []KeyValuePair `json:"tags,omitempty"`
}

func validateGeoPoint(g GeoPoint) error {
	if g.Lat == 0 || g.Lon == 0 {
		return errors.New("invalid lat or lon")
	}
	if g.Address.Prefecture == "" || g.Address.City == "" || g.Address.District == "" {
		return errors.New("invalid address")
	}
	return nil
}
