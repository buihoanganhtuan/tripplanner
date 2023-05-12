package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/buihoanganhtuan/tripplanner/backend/web_service/planner"
	"github.com/gorilla/mux"
)

type GraphError []planner.PointId
type CycleError []GraphError
type MultiFirstError GraphError
type MultiLastError GraphError
type SimulFirstAndLastError GraphError
type UnknownNodeIdError GraphError
type PointOrder []planner.PointId
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
	eh = planner.ErrorHandler(func(w http.ResponseWriter, r *http.Request) (error, planner.ErrorResponse) {
		id := mux.Vars(r)["id"]

		ctx := context.Background()
		tx, err := cst.Db.BeginTx(ctx, nil)
		if err != nil {
			return err, utils.NewDatabaseQueryError()
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
			return err, utils.NewDatabaseQueryError()
		}

		var ed, cd, lastModified time.Time
		var expectedDate, createdDate *time.Time

		if !anonymous {
			ed, err = time.Parse(cst.DatetimeFormat, edStr)
			if err != nil {
				return err, utils.NewServerParseError()
			}
			expectedDate = &ed

			cd, err = time.Parse(cst.DatetimeFormat, cdStr)
			if err != nil {
				return err, utils.NewServerParseError()
			}
			createdDate = &cd
		}

		lastModified, err = time.Parse(cst.DatetimeFormat, lmStr)
		if err != nil {
			return err, utils.NewServerParseError()
		}

		var edges []Edge
		var rows *sql.Rows
		rows, err = tx.Query("select PointId, NextPointId, Start, DurationHr, DurationMin, CostAmount, CostUnit, Transport from ? order by ord where tripId = ?", cst.SqlEdgeTableVar, id)
		if err != nil {
			return err, utils.NewDatabaseQueryError()
		}

		for rows.Next() {
			var pointId, nextPointId, _start, costUnit, transport string
			var durationHr, durationMin, costAmount int
			err = rows.Scan(&pointId, &nextPointId, &_start, &durationHr, &durationMin, &costAmount, &costUnit, &transport)
			if err != nil {
				return err, utils.NewDatabaseQueryError()
			}

			var start time.Time
			start, err = time.Parse(cst.DatetimeFormat, _start)
			if err != nil {
				return err, utils.NewServerParseError()
			}

			edges = append(edges, Edge{
				PointId:     planner.PointId(pointId),
				NextPointId: planner.PointId(nextPointId),
				Start:       utils.JsonDateTime(start),
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
			return err, utils.NewDatabaseQueryError()
		}

		var body []byte
		tp := "registered"
		if anonymous {
			tp = "anonymous"
		}
		body, err = json.Marshal(planner.Trip{
			Id:           planner.TripId(id),
			UserId:       uid,
			Type:         tp,
			Name:         name,
			DateExpected: (*utils.JsonDateTime)(expectedDate),
			DateCreated:  (*utils.JsonDateTime)(createdDate),
			LastModified: utils.JsonDateTime(lastModified),
			Budget: planner.Cost{
				Amount: budget,
				Unit:   budgetUnit,
			},
			PreferredMode: tm,
			PlanResult:    edges,
		})

		if err != nil {
			return err, utils.NewUnknownError()
		}

		w.Write(body)
		w.WriteHeader(http.StatusOK)
		return nil, planner.ErrorResponse{}
	})
	return eh
}

func newPlanTripHandler(anonymous bool) utils.ErrorHandler {
	return utils.ErrorHandler(func(w http.ResponseWriter, r *http.Request) (error, planner.ErrorResponse) {
		tripId, present := mux.Vars(r)["id"]
		if !present {
			return nil, utils.NewInvalidIdError()
		}

		var expectDateStr, budgetStr, budgetUnitStr, transportStr string
		var err error
		row := cst.Db.QueryRow("select ExpectedDate, BudgetLimit, BudgetUnit, PreferredTransportMode from ? where TripId = ?", cst.Ev.Var(cst.SqlTripTableVar), tripId)
		if err = row.Scan(&expectDateStr, &budgetStr, &budgetUnitStr, &transportStr); err == sql.ErrNoRows {
			return err, utils.NewInvalidIdError()
		}

		var expectDate time.Time
		expectDate, err = time.Parse(cst.DatetimeFormat, expectDateStr)
		if err != nil {
			return err, utils.NewServerParseError()
		}

		var budget int
		budget, err = strconv.Atoi(budgetStr)
		if err != nil {
			return err, utils.NewServerParseError()
		}

		allowedUnits := utils.NewSet[string]("JPY", "USD")
		allowedModes := utils.NewSet[string]("train", "bus", "walk")
		if !allowedUnits.Contains(budgetUnitStr) || allowedModes.Contains(transportStr) {
			return errors.New(""), utils.NewServerParseError()
		}

		// Get the points in the trip
		var rows *sql.Rows
		rows, err = cst.Db.Query("select Id, GeoPointId, Name, Lat, Lon from ? where TripId = ?", cst.Ev.Var(cst.SqlPointTableVar), tripId)
		if err != nil {
			return err, utils.NewDatabaseQueryError()
		}

		points := map[PointId]Point{}
		for rows.Next() {
			var pid PointId
			var gpid GeoPointId
			var name string
			var lat, lon float64
			err = rows.Scan(&pid, &gpid, &name, &lat, &lon)
			if err != nil {
				return err, utils.NewUnknownError()
			}
			points[pid] = Point{
				Id:         pid,
				Name:       name,
				GeoPointId: gpid,
				Lat:        lat,
				Lon:        lon,
			}
		}

		if rows.Err() != nil {
			return rows.Err(), utils.NewDatabaseQueryError()
		}

		// Get the point constraints and construct edges
		rows, err = cst.Db.Query("select FirstPointId, SecondPointId, ConstraintType from ? where FirstPointId in ("+strings.Join(pids.Values(), ",")+")", cst.PointConstraintTable)
		if err != nil {
			return err, utils.NewDatabaseQueryError()
		}

		for rows.Next() {
			var pid1, pid2 PointId
			var t string
			rows.Scan(&pid1, &pid2, &t)
			p1, ok := points[pid1]
			if !ok {
				return fmt.Errorf("unknown point id %s", pid1), utils.NewServerParseError()
			}
			p2, ok := points[pid2]
			if !ok {
				return fmt.Errorf("unknown point id %s", pid1), utils.NewServerParseError()
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
				return fmt.Errorf("unknown constraint type %s", t), utils.NewServerParseError()
			}
		}

		// Get the candidate orders based on constraints
		var candTrips []PointOrder
		maxTrips := cst.MaxCandidateTrips
		if anonymous {
			maxTrips = cst.MaxCandidateAnonTrips
		}
		candTrips, err = topologicalSort(utils.GetMapValues[PointId, Point](points), maxTrips)

		if err != nil {
			switch err := err.(type) {
			case CycleError:
				var errs []utils.ErrorDescriptor
				for _, ge := range []GraphError(err) {
					var cycle []string
					for _, pid := range []PointId(ge) {
						cycle = append(cycle, points[pid].Name)
					}
					errs = append(errs, utils.ErrorDescriptor{
						Domain:  cst.WebServiceName,
						Reason:  "Cycle found in graph",
						Message: strings.Join(cycle, ","),
					})
				}
				return err, planner.ErrorResponse{
					Code:    http.StatusInternalServerError,
					Message: "Nodes form cycle(s)",
					Errors:  errs,
				}
			case MultiFirstError, MultiLastError, SimulFirstAndLastError, UnknownNodeIdError:
				return err, planner.ErrorResponse{
					Code: http.StatusInternalServerError,
					// TODO
				}
			default:
				panic(err)
			}
		}

		// Get all close routes for each point
		var nearbyRoutePoints map[PointId][]GeoPoint
		for _, p := range utils.GetMapValues[PointId, Point](points) {
			var nps []GeoPoint
			nps, err = getNearbyRoutePoints(p, cst.RouteSearchRadius)
			if err != nil {
				return err, utils.NewUnknownError()
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

		return nil, planner.ErrorResponse{}
	})
}

/*
Extract possible solutions to a certain DAG ordering. As the number of solutions can be
quite large, we terminate the search when the number of results found thus far exceed lim
*/
func topologicalSort(points []Point, startTime time.Time, lim int) ([]PointOrder, error) {
	intIds := map[PointId]int{}
	pointIds := map[int]PointId{}

	mapback := func(intIds []int) []PointId {
		var ids []PointId
		for _, id := range intIds {
			ids = append(ids, pointIds[id])
		}
		return ids
	}

	// convert pointId (string) to integer id
	var first, last, both []int
	for i, p := range points {
		intIds[p.Id] = i
		pointIds[i] = p.Id
		if p.First && p.Last {
			both = append(both, i)
		}
		if p.First {
			first = append(first, i)
		}
		if p.Last {
			last = append(last, i)
		}
	}

	if len(first) > 1 {
		return nil, MultiFirstError(mapback(first))
	}

	if len(last) > 1 {
		return nil, MultiLastError(mapback(last))
	}

	if len(both) > 0 {
		return nil, MultiLastError(mapback(both))
	}

	// verify that there's no unknown node
	var unknown []PointId
	for _, p := range points {
		if _, ok := intIds[p.Id]; ok {
			continue
		}
		unknown = append(unknown, p.Id)
	}
	if len(unknown) > 0 {
		return nil, UnknownNodeIdError(unknown)
	}

	// construct the directed edges, prepare the in-degree count for each node
	indeg := make([]int, len(intIds))
	adj := make([][]int, len(intIds))
	for i, p := range points {
		for _, next := range p.Next {
			j := intIds[next]
			indeg[j]++
			adj[i] = append(adj[j], i)
		}
	}

	// Check for cycle
	cycles := GetCycles(indeg, adj)
	if cycles != nil {
		var ce []GraphError
		for _, c := range cycles {
			ce = append(ce, mapback(c))
		}
		return nil, CycleError(ce)
	}

	sortFn := func(i, j int) bool {
		d1 := points[i].Arrival.Defined
		d2 := points[j].Arrival.Defined
		if d1 && d2 {
			t1 := time.Time(points[i].Arrival.Value.before)
			t2 := time.Time(points[j].Arrival.Value.before)
			return t1.Before(t2)
		}
		if !d1 && !d2 || !d1 && d2 {
			return false
		}
		return true
	}

	var res []PointOrder
	var dfs func([]int, []int, time.Time)
	dfs = func(q, cur []int, t time.Time) {
		if len(res) >= lim {
			return
		}
		if len(q) == 0 {
			res = append(res, PointOrder(mapback(cur)))
			return
		}

		var qc []int
		qc = append(qc, q...)
		sort.SliceStable(qc, sortFn) // prioritize points with deadline first
		if points[qc[0]].Arrival.Defined && t.After(time.Time(points[qc[0]].Arrival.Value.before)) {
			return
		}

		// backtrack
		for i := 0; i < len(qc); i++ {
			tmp := qc[i]
			cur = append(cur, tmp)
			qc[i] = qc[len(qc)-1]
			qc = qc[:len(qc)-1]
			var added int
			for _, j := range adj[tmp] {
				indeg[j]--
				if indeg[j] == 0 {
					added++
					qc = append(qc, j)
				}
			}

			if points[qc[i]].Duration.Defined {
				dfs(qc, cur, t.Add(points[qc[i]].Duration.Value))
			} else {
				dfs(qc, cur, t)
			}

			for _, j := range adj[tmp] {
				indeg[j]++
			}
			qc = qc[:len(qc)-added]
			qc = append(qc, qc[i])
			qc[i] = tmp
			cur = cur[:len(cur)-1]
		}
	}

	var q, cur []int
	for i, d := range indeg {
		if d == 0 {
			q = append(q, i)
		}
	}
	dfs(q, cur, startTime)

	return res, nil
}

func GetCycles(indeg []int, adj [][]int) Cycles {
	var indegCp []int
	indegCp = append(indegCp, indeg...)

	var st []int
	for i, d := range indegCp {
		if d == 0 {
			st = append(st, i)
		}
	}

	for len(st) > 0 {
		i := st[len(st)-1]
		st = st[:len(st)-1]
		for _, j := range adj[i] {
			indegCp[j]--
			if indegCp[j] == 0 {
				st = append(st, j)
			}
		}
	}

	ok := true
	for _, d := range indegCp {
		if d > 0 {
			ok = false
		}
	}

	if ok {
		return nil
	}

	var res Cycles

	visited := make([]bool, len(adj))
	inStack := make([]bool, len(adj))
	var path []int

	var dfs func(int)
	dfs = func(i int) {
		if visited[i] {
			return
		}
		visited[i] = true
		inStack[i] = true
		path = append(path, i)
		for _, j := range adj[i] {
			if inStack[j] {
				// Found cycle
				var c []int
				var add bool
				for _, node := range path {
					add = add || node == j
					if add {
						c = append(c, node)
					}
				}
				res = append(res, c)
			}
			dfs(j)
		}
		path = path[:len(path)-1]
		inStack[i] = false
	}

	for i := 0; i < len(adj); i++ {
		dfs(i)
	}
	return res
}

/*
Get latitude and longitude of bottom left and top right of a square
centered at current position and has side = 2*dist
*/
func getNearbyRoutePoints(point Point, dist float64) ([]GeoPoint, error) {

	latBits := cst.GeohashLen / 2
	lonBits := cst.GeohashLen/2 + cst.GeohashLen%2
	numLats := int64(1) << int64(latBits)
	numLons := int64(1) << int64(lonBits)
	dlat := float64(180) / float64(numLats)
	dlon := float64(360) / float64(numLons)

	var lo, hi int
	hi = int(point.Lat / dlat)
	for lo < hi {
		mid := (lo + hi + 1) / 2
		lat := float64(-90) + float64(mid)*dlat
		if haversine(point.Lat, point.Lon, lat, point.Lon) >= dist {
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	blat := int64(lo)

	lo = int(point.Lat/dlat) + 1
	hi = 1<<latBits - 1
	for lo < hi {
		mid := (lo + hi) / 2
		lat := float64(-90) + float64(mid)*dlat
		if haversine(point.Lat, point.Lon, lat, point.Lon) >= dist {
			hi = mid
		} else {
			lo = mid + 1
		}
	}
	tlat := int64(lo)

	lo = 0
	hi = int(point.Lon / dlon)
	for lo < hi {
		mid := (lo + hi + 1) / 2
		lon := float64(-180) + float64(mid)*dlon
		if haversine(point.Lat, point.Lon, point.Lat, lon) >= dist {
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	llon := int64(lo)

	lo = int(point.Lon/dlon) + 1
	hi = 1<<lonBits - 1
	for lo < hi {
		mid := (lo + hi) / 2
		lon := float64(-180) + float64(mid)*dlon
		if haversine(point.Lat, point.Lon, point.Lat, lon) >= dist {
			hi = mid
		} else {
			lo = mid + 1
		}
	}
	rlon := int64(lo)

	var geoHashes []string
	for j := blat; j <= tlat; j++ {
		for k := llon; k <= rlon; k++ {
			geoHashes = append(geoHashes, strconv.FormatInt(j+k<<latBits, 10))
		}
	}

	q := fmt.Sprintf(`select RouteId, GeoPointId, NodeLat, NodeLon 
						from %s 
						where GeoHashId in (?)`, cst.SqlWayTableVar)
	rows, err := cst.Db.Query(q, strings.Join(geoHashes, ","))
	if err != nil {
		return nil, err
	}

	var tmp map[int64]GeoPoint
	for rows.Next() {
		var rid, gid int64
		var latStr, lonStr string
		if err = rows.Scan(&rid, &gid, &latStr, &lonStr); err != nil {
			return nil, err
		}
		var lat, lon float64
		lat, err = strconv.ParseFloat(latStr, 64)
		if err != nil {
			return nil, err
		}
		lon, err = strconv.ParseFloat(lonStr, 64)
		if err != nil {
			return nil, err
		}
		if haversine(point.Lat, point.Lon, lat, lon) > dist {
			continue
		}
		gp := tmp[gid]
		gp.Routes = append(gp.Routes, RouteId(rid))
		tmp[gid] = gp
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	var res []GeoPoint
	for _, gp := range tmp {
		res = append(res, gp)
	}

	return res, nil
}

func FindShortestPath(src utils.Set[RouteId], dst utils.Set[RouteId], preferMode string, budget int) []RouteId {

}

func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	r := 6378.137e3
	a := math.Sin((lat2 - lat1) / 2)
	b := math.Sin((lon2 - lon1) / 2)
	return 2 * r * math.Asin(math.Sqrt(a*a+math.Cos(lat1)*math.Cos(lat2)*b*b))
}
