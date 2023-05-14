package app

import "github.com/buihoanganhtuan/tripplanner/backend/web_service/domain"

// Application implementation of domain's types
type Point struct {
	Id         domain.PointId          `json:"id"`
	TripId     domain.TripId           `json:"tripId"`
	GeoPointId domain.GeoPointId       `json:"geoPointId"`
	Arrival    *PointArrivalConstraint `json:"arrivalConstraint,omitempty"`
	Duration   Duration                `json:"duration"`
	Before     *PointBeforeConstraint  `json:"beforeConstraint,omitempty"`
	After      *PointAfterConstraint   `json:"afterConstraint,omitempty"`
	First      bool                    `json:"first"`
	Last       bool                    `json:"last"`
}

type PointArrivalConstraint struct {
	Before DateTime `json:"before"`
}

type PointAfterConstraint struct {
	Points []domain.PointId `json:"points"`
}

type PointBeforeConstraint struct {
	Points []domain.PointId `json:"points"`
}

type GeoPoint struct {
	Id      domain.GeoPointId `json:"id"`
	Lat     float64           `json:"lat"`
	Lon     float64           `json:"lon"`
	Name    *string           `json:"name,omitempty"`
	Address Address           `json:"address"`
	Tags    []KeyValuePair    `json:"tags,omitempty"`
}

// ************************************************************
