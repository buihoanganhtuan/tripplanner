package domain

import "time"

/*
A domain-level package Define domain types that model
real entities involved in the business, independent
from any underlying technology.
*/

type Repository interface {
	CreateTransaction() (TransactionId, error)
	CommitTransaction(id TransactionId) error
	RollbackTransaction(id TransactionId) error

	User(id UserId, tid TransactionId) (User, error)
	CreateUser(u User, tid TransactionId) (User, error)
	UpdateUser(u User, tid TransactionId) (User, error)
	DeleteUser(id UserId, tid TransactionId) error
	GetUserTrips(id UserId, tid TransactionId) ([]Trip, error)

	GeoPoint(id GeoPointId) (GeoPoint, error)
	GeoPoints(ids []GeoPointId) ([]GeoPoint, error)
	GeoPointsWithHashes(hs []GeoHashId) ([]GeoPoint, error)

	Point(id PointId) (Point, error)
	Points(ids []PointId) ([]Point, error)
	PointsWithTrip(id TripId) ([]Point, error)

	GetTrip(id TripId, tid TransactionId) (Trip, error)
	AddTrip(t Trip, tid TransactionId) (Trip, error)
	DeleteTrip(id TripId, tid TransactionId) error
}

type Api interface {
	GetUser(id UserId) (User, error)
	CreateUser(u User) (User, error)
	UpdateUser(u User) (User, error)
	ListUsers() ([]User, error)
	DeleteUser(id UserId) error
}

type Domain struct {
	repo Repository
	api  Api
}

type TransactionId string

type DateTime time.Time

func (dt DateTime) before(odt DateTime) bool {
	return time.Time(dt).Before(time.Time(odt))
}

func (dt DateTime) after(odt DateTime) bool {
	return time.Time(dt).After(time.Time(odt))
}

func (dt DateTime) add(d Duration) DateTime {
	var dur time.Duration
	switch d.Unit {
	case "hour":
		dur = time.Duration(d.Len * int(time.Hour))
	case "min":
		dur = time.Duration(d.Len * int(time.Minute))
	default:
		panic("unknown duration unit " + d.Unit)
	}
	return DateTime(time.Time(dt).Add(dur))
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

type Cost struct {
	Amount int    `json:"amount"`
	Unit   string `json:"unit"`
}

type Duration struct {
	Len  int    `json:"duration"`
	Unit string `json:"unit"`
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
