package domain

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
