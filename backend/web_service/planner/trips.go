package planner

type Trip struct {
	Id            TripId    `json:"id"`
	Type          string    `json:"type"`
	UserId        string    `json:"userId,omitempty"`
	Name          string    `json:"name,omitempty"`
	DateExpected  *DateTime `json:"dateExpected,omitempty"`
	DateCreated   *DateTime `json:"dateCreated,omitempty"`
	LastModified  *DateTime `json:"lastModified,omitempty"`
	Budget        Cost      `json:"budgetLimit"`
	PreferredMode string    `json:"preferredTransportMode"`
	PlanResult    []Path    `json:"planResult"`
}

type TripService interface {
	GetTrip(id TripId) (Trip, error)
	CreateTrip(t Trip) (Trip, error)
	ListTrips() ([]Trip, error)
	UpdateTrip(t Trip) (Trip, error)
	ReplaceTrip(t Trip) (Trip, error)
	PlanTrip(id TripId) (Trip, error)
	DeleteTrip(id TripId) error
}

type TripId string

type DateTime struct {
	Year   int    `json:"year"`
	Month  int    `json:"month"`
	Day    int    `json:"day"`
	Hour   int    `json:"hour"`
	Minute int    `json:"min"`
	Zone   string `json:"timezone"`
}

type Cost struct {
	Amount int    `json:"amount"`
	Unit   string `json:"unit"`
}

type Duration struct {
	Duration int    `json:"duration"`
	Unit     string `json:"unit"`
}

type Path struct {
	PointId     PointId         `json:"pointId"`
	NextPointId PointId         `json:"nextPointId"`
	Start       DateTime        `json:"start"`
	Duration    Duration        `json:"duration"`
	Transports  []TransportInfo `json:"transports"`
}

type TransportInfo struct {
	Start    DateTime `json:"start"`
	Duration Duration `json:"duration"`
	Type     string   `json:"type"`
	Info     any      `json:"info"`
}

type BusInfo struct {
	Cost      Cost   `json:"cost"`
	Operator  string `json:"operator"`
	Route     string `json:"route"`
	BusNumber string `json:"busNumber"`
}

type TrainInfo struct {
	Cost     Cost   `json:"cost"`
	Operator string `json:"operator"`
	Line     string `json:"line"`
}

type WalkInfo struct {
}
