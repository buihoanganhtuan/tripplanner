package domain

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
