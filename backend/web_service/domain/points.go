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

type PointService interface {
	GetPoint(id PointId) (Point, error)
	CreatePoint(p Point) (Point, error)
	ListPoints() ([]Point, error)
	UpdatePoint(p Point) (Point, error)
	ReplacePoint(p Point) (Point, error)
	DeletePoint(id PointId) error

	GetGeoPoint(id GeoPointId) (GeoPoint, error)
	GetNearbyPoints(id GeoPointId, dist float64) ([]GeoPoint, error)
}

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
