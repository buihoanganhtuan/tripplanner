package osm

import (
	"database/sql"
	"encoding/json"
	"os"

	"github.com/buihoanganhtuan/tripplanner/backend/web_service/domain"
)

type Odpt struct {
	Domain       domain.Domain
	GeospatialDb *sql.DB
}

type OdptJsonLdCommon struct {
	Context    string  `json:"@context"`
	Ucode      string  `json:"@id"`
	Type       string  `json:"@type"`
	Id         string  `json:"owl:sameAs"`
	Date       string  `json:"dc:date"`
	ValidUntil *string `json:"dct:valid,omitempty"`
}

type OdptBus struct {
	OdptJsonLdCommon
	BusNumber          string   `json:"odpt:busNumber"`
	StatusUpdateFreq   int      `json:"odpt:frequency"`
	BusRouteId         string   `json:"odpt:busroutePattern"`
	BusTimeTableId     *string  `json:"odpt:busTimetable,omitempty"`
	OperatorId         string   `json:"odpt:operator"`
	FirstBusStopId     *string  `json:"odpt:startingBusstopPole,omitempty"`
	LastBusStopId      *string  `json:"odpt:terminalBusstopPole,omitempty"`
	CurrentBusStopId   *string  `json:"odpt:fromBusstopPole,omitempty"`
	CurrentBusStopTime *string  `json:"odpt:fromBusstopPoleTime,omitempty"`
	NextBusStopId      *string  `json:"odpt:toBusstopPole,omitempty"`
	Progress           *float64 `json:"odpt:progress,omitempty"`
	Longitude          *float64 `json:"geo:long,omitempty"`
	Latitude           *float64 `json:"geo:lat,omitempty"`
	Speed              *float64 `json:"odpt:speed,omitempty"`
	Azimuth            *float64 `json:"odpt:azimuth,omitempty"`
	DoorStatus         *string  `json:"odpt:doorStatus,omitempty"`
	OccupancyStatus    *string  `json:"odpt:occupancyStatus,omitempty"`
}

type OdptBusTimeTable struct {
	OdptJsonLdCommon
	RevisionDate       *string              `json:"dct:issued,omitempty"`
	BusLineNameJap     *string              `json:"dc:title,omitempty"`
	BusLineNameKana    *string              `json:"odpt:kana,omitempty"`
	OperatorId         string               `json:"odpt:operator"`
	BusRouteId         string               `json:"odpt:busroutePattern"`
	CalendarId         string               `json:"odpt:calendar"`
	BusTimeTableObject []BusTimeTableObject `json:"odpt:busTimetableObject"`
}

type BusTimeTableObject struct {
	BusStopOrder    int     `json:"odpt:index"`
	BusStopId       string  `json:"odpt:busstopPole"`
	ArrivalTime     *string `json:"odpt:arrivalTime,omitempty"`
	DepartureTime   *string `json:"odpt:departureTime,omitempty"`
	DestinationSign *string `json:"odpt:destinationSign,omitempty"`
	IsFlatFare      *bool   `json:"odpt:isNonStepBus,omitempty"`
	IsLateNight     *bool   `json:"odpt:isMidnight,omitempty"`
	CanEmbark       *bool   `json:"odpt:canGetOn,omitempty"`
	CanDisembark    *bool   `json:"odpt:canGetOff,omitempty"`
	Note            *string `json:"odpt:note,omitempty"`
}

type OdptBusStop struct {
	OdptJsonLdCommon
	JapTitle           string           `json:"dc:title"`
	Kana               *string          `json:"odpt:kana,omitempty"`
	MultilingualTitles *json.RawMessage `json:"title,omitempty"`
	Longitude          *float64         `json:"geo:long,omitempty"`
	Latitude           *float64         `json:"geo:lat,omitempty"`
	Region             *json.RawMessage `json:"ug:region,omitempty"`
	BusRouteIds        *[]string        `json:"odpt:busroutePattern,omitempty"`
	OperatorIds        []string         `json:"odpt:operator"`
	PoleNumber         *string          `json:"odpt:busstopPoleNumber,omitempty"`
	TimetableIds       *[]string        `json:"odpt:busstopPoleTimetable,omitempty"`
}

type OdptBusRoute struct {
	OdptJsonLdCommon
	JapTitle       string             `json:"dc:title"`
	Kana           *string            `json:"odpt:kana,omitempty"`
	OperatorId     string             `json:"odpt:operator"`
	RouteId        *string            `json:"odpt:busroute,omitempty"`
	RoutePattern   *string            `json:"odpt:pattern,omitempty"`
	Direction      *string            `json:"odpt:direction,omitempty"`
	Region         *json.RawMessage   `json:"ug:region,omitempty"`
	BusStopOrder   []OdptBusStopOrder `json:"odpt:busstopPoleOrder"`
	Note           *string            `json:"odpt:note,omitempty"`
	BusLocationUrl *string            `json:"odpt:busLocationURL,omitempty"`
}

type OdptBusStopOrder struct {
	BusStopId     string  `json:"odpt:busstopPole"`
	BusStopIndex  int     `json:"odpt:index"`
	EmbarkDoor    *string `json:"odpt:openingDoorsToGetOn,omitempty"`
	DisembarkDoor *string `json:"odpt:openingDoorsToGetOff,omitempty"`
	Note          *string `json:"odpt:note,omitempty"`
}

type OdptBusRouteFare struct {
	OdptJsonLdCommon
	RevisionDate       *string `json:"dct:issued,omitempty"`
	OperatorId         string  `json:"odpt:operator"`
	EmbarkRouteId      string  `json:"odpt:fromBusroutePattern"`
	EmbarkStopIndex    int     `json:"odpt:fromBusstopPoleOrder"`
	EmbarkStopId       string  `json:"odpt:fromBusstopPole"`
	DisembarkRouteId   string  `json:"odpt:toBusroutePattern"`
	DisembarkStopIndex int     `json:"odpt:toBusstopPoleOrder"`
	DisembarkStopId    string  `json:"odpt:toBusstopPole"`
	FareYen            int     `json:"odpt:ticketFare"`
	ChildFareYen       *int    `json:"odpt:childTicketFare,omitempty"`
	IcFareYen          *int    `json:"odpt:icCardFare,omitempty"`
	ChildIcFareYen     *int    `json:"odpt:childIcCardFare,omitempty"`
}

type OdptBusStopTimeTable struct {
	OdptJsonLdCommon
	RevisionDate    *string            `json:"dct:issued,omitempty"`
	RouteName       *string            `json:"dc:title,omitempty"`
	BusStopId       string             `json:"odpt:busstopPole"`
	BusDirection    []string           `json:"odpt:busDirection"`
	BusRoute        []string           `json:"odpt:busroute"`
	OperatorId      string             `json:"odpt:operator"`
	DateOfOperation string             `json:"odpt:calendar"`
	TimeTableObject *[]TimeTableObject `json:"odpt:busstopPoleTimetableObject,omitempty"`
}

type TimeTableObject struct {
	ArrivalTime       *string `json:"odpt:arrivalTime,omitempty"`
	DepartureTime     string  `json:"odpt:departureTime"`
	DestinationStopId *string `json:"odpt:destinationBusstopPole,omitempty"`
	DestinationSign   *string `json:"odpt:destinationSign,omitempty"`
	BusRouteId        *string `json:"odpt:busroutePattern,omitempty"`
	BusStopOrder      *int    `json:"odpt:busroutePatternOrder,omitempty"`
	IsFlatFare        *bool   `json:"odpt:isNonStepBus,omitempty"`
	IsLateNight       *bool   `json:"odpt:isMidnight,omitempty"`
	CanEmbark         *bool   `json:"odpt:canGetOn,omitempty"`
	CanDisembark      *bool   `json:"odpt:canGetOff,omitempty"`
	Note              *string `json:"odpt:note,omitempty"`
}

type OdptRailDirection struct {
	OdptJsonLdCommon
	TravelDirectionJap          *string          `json:"dc:title,omitempty"`
	TravelDirectionMultilingual *json.RawMessage `json:"odpt:railDirectionTitle,omitempty"`
}

type OdptTrainLine struct {
	OdptJsonLdCommon
	LineNameJap          string           `json:"dc:title"`
	LineNameMultilingual *json.RawMessage `json:"odpt:railwayTitle,omitempty"`
	LineNameKana         *string          `json:"odpt:kana,omitempty"`
	OperatorId           string           `json:"odpt:operator"`
	LineCode             *string          `json:"odpt:lineCode,omitempty"`
	LineColor            *string          `json:"odpt:color,omitempty"`
	Region               *json.RawMessage `json:"ug:region,omitempty"`
	AscendDirectionId    *string          `json:"odpt:ascendingRailDirection,omitempty"`
	DescendDirectionId   *string          `json:"odpt:descendingRailDirection,omitempty"`
	StationOrder         []StationOrder   `json:"odpt:stationOrder"`
}

type StationOrder struct {
	StationId               string           `json:"odpt:station"`
	StationNameMultilingual *json.RawMessage `json:"odpt:stationTitle,omitempty"`
	StationIndex            int              `json:"odpt:index"`
}

type OdptStation struct {
	OdptJsonLdCommon
	StationNameJap          *string          `json:"dc:title,omitempty"`
	StationNameMultilingual *json.RawMessage `json:"odpt:stationTitle,omitempty"`
	OperatorId              string           `json:"odpt:operator"`
	TrainLineId             string           `json:"odpt:railway"`
	StationCode             *string          `json:"odpt:stationCode,omitempty"`
	Longitude               float64          `json:"geo:long,omitempty"`
	Latitude                float64          `json:"geo:lat,omitempty"`
	Region                  *json.RawMessage `json:"ug:region,omitempty"`
	StationEntranceIds      *[]string        `json:"odpt:exit,omitempty"`
	TransferableLineIds     *[]string        `json:"odpt:connectingRailway,omitempty"`
	StationTimeTableIds     *[]string        `json:"odpt:stationTimetable,omitempty"`
	PassengerSurveyIds      *[]string        `json:"odpt:passengerSurvey,omitempty"`
}

type OdptStationTimeTable struct {
	OdptJsonLdCommon
	RevisionDate            *string                  `json:"dct:issued,omitempty"`
	OperatorId              string                   `json:"odpt:operator"`
	LineId                  string                   `json:"odpt:railway"`
	LineNameMultilingual    *json.RawMessage         `json:"odpt:railwayTitle,omitempty"`
	StationId               *string                  `json:"odpt:station,omitempty"`
	StationNameMultilingual *json.RawMessage         `json:"odpt:stationTitle,omitempty"`
	DirectionId             *string                  `json:"odpt:railDirection,omitempty"`
	CalendarId              *string                  `json:"odpt:calendar,omitempty"`
	StationTimeTableObject  []StationTimeTableObject `json:"odpt:stationTimetableObject"`
	NoteMultilingual        *json.RawMessage         `json:"odpt:note,omitempty"`
}

type StationTimeTableObject struct {
	ArrivalTime               *string          `json:"odpt:arrivalTime,omitempty"`
	DepartureTime             *string          `json:"odpt:departureTime,omitempty"`
	OriginStationIds          *[]string        `json:"odpt:originStation,omitempty"`
	DestinationStationIds     *[]string        `json:"odpt:destinationStation,omitempty"`
	TransferStationIds        *[]string        `json:"odpt:viaStation,omitempty"`
	TransferLineIds           *[]string        `json:"odpt:viaRailway,omitempty"`
	TrainId                   *string          `json:"odpt:train,omitempty"`
	TrainNumber               *string          `json:"odpt:trainNumber,omitempty"`
	TrainType                 *string          `json:"odpt:trainType,omitempty"`
	TrainName                 *string          `json:"odpt:trainName,omitempty"`
	OperatorId                *string          `json:"odpt:trainOwner,omitempty"`
	IsLastTrain               *bool            `json:"odpt:isLast,omitempty"`
	IsOriginStation           *bool            `json:"odpt:isOrigin,omitempty"`
	ArrivalPlatformNumber     *string          `json:"odpt:platformNumber,omitempty"`
	PlatformNamesMultilingual *json.RawMessage `json:"odpt:platformName,omitempty"`
	NumCars                   *int             `json:"odpt:carComposition,omitempty"`
	NoteMultilingual          *json.RawMessage `json:"odpt:note,omitempty"`
}

type OdptTrain struct {
	OdptJsonLdCommon
	OperatorId            string             `json:"odpt:operator,omitempty"`
	LineId                string             `json:"odpt:railway"`
	DirectionId           *string            `json:"odpt:railDirection,omitempty"`
	TrainNumber           string             `json:"odpt:trainNumber"`
	TrainType             *string            `json:"odpt:trainType,omitempty"`
	TrainNameMultilingual *[]json.RawMessage `json:"odpt:trainName,omitempty"`
	LatestStationId       *string            `json:"odpt:fromStation,omitempty"`
	NextStationId         *string            `json:"odpt:toStation,omitempty"`
	OriginStationIds      *[]string          `json:"odpt:originStation,omitempty"`
	FinalStationIds       *[]string          `json:"odpt:destinationStation,omitempty"`
	TransferStationIds    *[]string          `json:"odpt:viaStation,omitempty"`
	TransferLineIds       *[]string          `json:"odpt:viaRailway,omitempty"`
	TrainOwnerId          *string            `json:"odpt:trainOwner,omitempty"`
	TrainOrder            *int               `json:"odpt:index,omitempty"`
	DelaySeconds          *int               `json:"odpt:delay,omitempty"`
	NumCars               *int               `json:"odpt:carComposition,omitempty"`
	NoteMultilingual      *json.RawMessage   `json:"odpt:note,omitempty"`
}

// This function extracts the bus stops stored in the odpt bus stop json file
// and store them in database
func (o *Odpt) processBusStops(odptBusStopFile string) error {
	file, err := os.Open(odptBusStopFile)
	if err != nil {
		return err
	}
	dec := json.NewDecoder(file)
	// Read opening bracket
	if _, err = dec.Token(); err != nil {
		return err
	}
	for dec.More() {
		var busStop OdptBusStop
		if err = dec.Decode(&busStop); err != nil {
			return err
		}

	}

	// read closing bracket
	if _, err = dec.Token(); err != nil {
		return err
	}
	return nil
}
