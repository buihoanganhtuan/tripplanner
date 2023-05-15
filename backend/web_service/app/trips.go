package app

import (
	"context"
	"database/sql"
	gjson "encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/buihoanganhtuan/tripplanner/backend/web_service/datastructure"
	"github.com/buihoanganhtuan/tripplanner/backend/web_service/domain"
	json "github.com/buihoanganhtuan/tripplanner/backend/web_service/encoding/json"
	"github.com/gorilla/mux"
)

// Application implementation of domain's types
type Trip struct {
	Id            domain.TripId `json:"id"`
	Type          string        `json:"type"`
	UserId        string        `json:"userId,omitempty"`
	Name          string        `json:"name,omitempty"`
	DateExpected  *json.DateTime     `json:"dateExpected,omitempty"`
	DateCreated   *json.DateTime     `json:"dateCreated,omitempty"`
	LastModified  *json.DateTime     `json:"lastModified,omitempty"`
	Budget        Cost          `json:"budgetLimit"`
	PreferredMode string        `json:"preferredTransportMode"`
	PlanResult    []Path        `json:"planResult"`
}

type GraphError []domain.PointId
type CycleError []GraphError
type MultiFirstError GraphError
type MultiLastError GraphError
type SimulFirstAndLastError GraphError
type UnknownNodeIdError GraphError
type PointOrder []domain.PointId
type Cycle []int
type Cycles []Cycle

func (ge GraphError) Error() string {
	var pids []string
	for _, pid := range ge {
		pids = append(pids, string(pid))
	}
	return strings.Join(pids, ",")
}

func (mf MultiFirstError) Error() string {
	return GraphError(mf).Error()
}

func (ml MultiLastError) Error() string {
	return GraphError(ml).Error()
}

func (un UnknownNodeIdError) Error() string {
	return GraphError(un).Error()
}

func (ce CycleError) Error() string {
	ges := []GraphError(ce)
	var sb strings.Builder
	for _, ge := range ges {
		sb.WriteString(ge.Error())
		sb.WriteString("\\n")
	}
	return sb.String()
}

func (sm SimulFirstAndLastError) Error() string {
	return GraphError(sm).Error()
}

func NewGetTripHandler(anonymous bool) planner.ErrorHandler {
	var eh planner.ErrorHandler
	eh = planner.ErrorHandler(func(w http.ResponseWriter, r *http.Request) (error, ErrorResponse) {
		id := mux.Vars(r)["id"]

		ctx := context.Background()
		tx, err := .Db.BeginTx(ctx, nil)
		if err != nil {
			return err, NewDatabaseQueryError()
		}

		var budget int
		var uid, name, edStr, cdStr, lmStr, budgetUnit, tm string
		if anonymous {

		}
		if anonymous {
			r := tx.QueryRow("select Name, LastModified, BudgetLimit, BudgetUnit, PreferredTransportMode from ? where id = ?", cst.Ev.Var(cst.SqlAnonTripTableVar), id)
			err = r.Scan(&name, &lmStr, &budget, &budgetUnit, &tm)
		} else {
			r := tx.QueryRow("select UserId, Name, ExpectedDate, CreatedDate, LastModified, BudgetLimit, BudgetUnit, PreferredTransportMode from ? where id = ?", cst.Ev.Var(cst.SqlTripTableVar), id)
			err = r.Scan(&uid, &name, &edStr, &cdStr, &lmStr, &budget, &budgetUnit, &tm)
		}

		if err != nil {
			return err, NewDatabaseQueryError()
		}

		var ed, cd, lastModified time.Time
		var expectedDate, createdDate *time.Time

		if !anonymous {
			ed, err = time.Parse(cst.DatetimeFormat, edStr)
			if err != nil {
				return err, NewServerParseError()
			}
			expectedDate = &ed

			cd, err = time.Parse(cst.DatetimeFormat, cdStr)
			if err != nil {
				return err, NewServerParseError()
			}
			createdDate = &cd
		}

		lastModified, err = time.Parse(cst.DatetimeFormat, lmStr)
		if err != nil {
			return err, NewServerParseError()
		}

		var edges []Edge
		var rows *sql.Rows
		rows, err = tx.Query("select planner.PointId, Nextplanner.PointId, Start, DurationHr, DurationMin, CostAmount, CostUnit, Transport from ? order by ord where tripId = ?", cst.SqlEdgeTableVar, id)
		if err != nil {
			return err, NewDatabaseQueryError()
		}

		for rows.Next() {
			var PointId, nextPointId, _start, costUnit, transport string
			var durationHr, durationMin, costAmount int
			err = rows.Scan(&PointId, &nextPointId, &_start, &durationHr, &durationMin, &costAmount, &costUnit, &transport)
			if err != nil {
				return err, NewDatabaseQueryError()
			}

			var start time.Time
			start, err = time.Parse(cst.DatetimeFormat, _start)
			if err != nil {
				return err, NewServerParseError()
			}

			edges = append(edges, Edge{
				planner.PointId:     planner.PointId(PointId),
				Nextplanner.PointId: planner.PointId(nextPointId),
				Start:               json.DateTime(start),
				Duration: planner.Duration{
					Hour: durationHr,
					Min:  durationMin,
				},
				Cost: planner.Cost{
					Amount: costAmount,
					Unit:   costUnit,
				},
				Transport: transport,
			})
		}

		err = rows.Err()
		if err != nil {
			return err, NewDatabaseQueryError()
		}

		var body []byte
		tp := "registered"
		if anonymous {
			tp = "anonymous"
		}
		body, err = gjson.Marshal(planner.Trip{
			Id:           planner.TripId(id),
			UserId:       uid,
			Type:         tp,
			Name:         name,
			DateExpected: (*json.DateTime)(expectedDate),
			DateCreated:  (*json.DateTime)(createdDate),
			LastModified: json.DateTime(lastModified),
			Budget: planner.Cost{
				Amount: budget,
				Unit:   budgetUnit,
			},
			PreferredMode: tm,
			PlanResult:    edges,
		})

		if err != nil {
			return err, NewUnknownError()
		}

		w.Write(body)
		w.WriteHeader(http.StatusOK)
		return nil, ErrorResponse{}
	})
	return eh
}

func newPlanTripHandler(anonymous bool) utils.ErrorHandler {
	return utils.ErrorHandler(func(w http.ResponseWriter, r *http.Request) (error, ErrorResponse) {
		tripId, present := mux.Vars(r)["id"]
		if !present {
			return nil, NewInvalidIdError()
		}

		var expectDateStr, budgetStr, budgetUnitStr, transportStr string
		var err error
		row := cst.Db.QueryRow("select ExpectedDate, BudgetLimit, BudgetUnit, PreferredTransportMode from ? where TripId = ?", cst.Ev.Var(cst.SqlTripTableVar), tripId)
		if err = row.Scan(&expectDateStr, &budgetStr, &budgetUnitStr, &transportStr); err == sql.ErrNoRows {
			return err, NewInvalidIdError()
		}

		var expectDate time.Time
		expectDate, err = time.Parse(cst.DatetimeFormat, expectDateStr)
		if err != nil {
			return err, NewServerParseError()
		}

		var budget int
		budget, err = strconv.Atoi(budgetStr)
		if err != nil {
			return err, NewServerParseError()
		}

		allowedUnits := datastructures.NewSet[string]("JPY", "USD")
		allowedModes := datastructures.NewSet[string]("train", "bus", "walk")
		if !allowedUnits.Contains(budgetUnitStr) || allowedModes.Contains(transportStr) {
			return errors.New(""), NewServerParseError()
		}

		// Get the points in the trip
		var rows *sql.Rows
		rows, err = cst.Db.Query("select Id, Geoplanner.PointId, Name, Lat, Lon from ? where TripId = ?", cst.Ev.Var(cst.SqlPointTableVar), tripId)
		if err != nil {
			return err, NewDatabaseQueryError()
		}

		points := datastructures.NewMap[planner.PointId, planner.Point]()
		for rows.Next() {
			var pid planner.PointId
			var gpid Geoplanner.PointId
			var name string
			var lat, lon float64
			err = rows.Scan(&pid, &gpid, &name, &lat, &lon)
			if err != nil {
				return err, NewUnknownError()
			}
			points.Put(pid, planner.Point{
				Id:         pid,
				Name:       name,
				GeoPointId: gpid,
				Lat:        lat,
				Lon:        lon,
			})
		}

		if rows.Err() != nil {
			return rows.Err(), NewDatabaseQueryError()
		}

		// Get the point constraints and construct edges
		rows, err = cst.Db.Query("select Firstplanner.PointId, Secondplanner.PointId, ConstraintType from ? where Firstplanner.PointId in ("+strings.Join(pids.Values(), ",")+")", cst.PointConstraintTable)
		if err != nil {
			return err, NewDatabaseQueryError()
		}

		for rows.Next() {
			var pid1, pid2 planner.PointId
			var t string
			rows.Scan(&pid1, &pid2, &t)
			p1, ok := points[pid1]
			if !ok {
				return fmt.Errorf("unknown point id %s", pid1), NewServerParseError()
			}
			p2, ok := points[pid2]
			if !ok {
				return fmt.Errorf("unknown point id %s", pid1), NewServerParseError()
			}
			switch t {
			case "before":
				p1.Next = append(p1.Next, pid2)
			case "after":
				p2.Next = append(p2.Next, pid1)
			case "first":
				p1.First = true
			case "last":
				p1.Last = true
			default:
				return fmt.Errorf("unknown constraint type %s", t), NewServerParseError()
			}
		}

		// Get the candidate orders based on constraints
		var candTrips []PointOrder
		maxTrips := cst.MaxCandidateTrips
		if anonymous {
			maxTrips = cst.MaxCandidateAnonTrips
		}
		candTrips, err = topologicalSort(points.Values(), maxTrips)

		if err != nil {
			switch err := err.(type) {
			case CycleError:
				var errs []ErrorDescriptor
				for _, ge := range []GraphError(err) {
					var cycle []string
					for _, pid := range []planner.PointId(ge) {
						cycle = append(cycle, points.Get(pid).Name)
					}
					errs = append(errs, ErrorDescriptor{
						Domain:  cst.WebServiceName,
						Reason:  "Cycle found in graph",
						Message: strings.Join(cycle, ","),
					})
				}
				return err, ErrorResponse{
					Code:    http.StatusInternalServerError,
					Message: "Nodes form cycle(s)",
					Errors:  errs,
				}
			case MultiFirstError, MultiLastError, SimulFirstAndLastError, UnknownNodeIdError:
				return err, ErrorResponse{
					Code: http.StatusInternalServerError,
					// TODO
				}
			default:
				panic(err)
			}
		}

		// Get all close routes for each point
		var nearbyRoutePoints datastructures.Map[planner.PointId, []planner.GeoPoint]
		for _, p := range points.Values() {
			var nps []GeoPoint
			nps, err = getNearbyRoutePoints(p, cst.RouteSearchRadius)
			if err != nil {
				return err, NewUnknownError()
			}
			nearbyRoutePoints[p.Id] = nps
		}

		// Fetch available routes for each edge (use internal osm data to get routes + external transport api based on date to get available transports)
		for _, ord := range candTrips {
			for i := 0; i < len(ord)-1; i++ {
				src := ord[i]
				dst := ord[i+1]
				srcRoutePoints := nearbyRoutePoints[src]
				dstRoutePoints := nearbyRoutePoints[dst]

			}
		}

		return nil, ErrorResponse{}
	})
}


