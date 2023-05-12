package planner

type Point struct {
	Id         PointId                 `json:"id"`
	TripId     TripId                  `json:"tripId"`
	GeoPointId GeoPointId              `json:"geoPointId"`
	Arrival    *PointArrivalConstraint `json:"arrivalConstraint,omitempty"`
	Duration   Duration                `json:"duration"`
	Before     *PointBeforeConstraint  `json:"beforeConstraint,omitempty"`
	After      *PointAfterConstraint   `json:"afterConstraint,omitempty"`
	First      bool                    `json:"first"`
	Last       bool                    `json:"last"`
}

type PointService interface {
	GetPoint(id PointId) (Point, error)
	CreatePoint(p Point) (Point, error)
	ListPoints() ([]Point, error)
	UpdatePoint(p Point) (Point, error)
	ReplacePoint(p Point) (Point, error)
	DeletePoint(id PointId) error
}

type PointId string

type PointArrivalConstraint struct {
	Before DateTime `json:"before"`
}

type PointAfterConstraint struct {
	Points []PointId `json:"points"`
}

type PointBeforeConstraint struct {
	Points []PointId `json:"points"`
}

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

type GeoPointId int64

type GeoPoint struct {
	Id      GeoPointId     `json:"id"`
	Lat     float64        `json:"lat"`
	Lon     float64        `json:"lon"`
	Name    *string        `json:"name,omitempty"`
	Address Address        `json:"address"`
	Tags    []KeyValuePair `json:"tags,omitempty"`
}
