package domain

/*
A domain-level package Define domain types that model
real entities involved in the business, independent
from any underlying technology.
*/

type DateTime struct {
	Year   int    `json:"year"`
	Month  int    `json:"month"`
	Day    int    `json:"day"`
	Hour   int    `json:"hour"`
	Minute int    `json:"min"`
	Zone   string `json:"timezone"`
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
