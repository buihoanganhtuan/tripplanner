package trips

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	cst "github.com/buihoanganhtuan/tripplanner/backend/web_service/_constants"
	utils "github.com/buihoanganhtuan/tripplanner/backend/web_service/_utils"
	"github.com/gorilla/mux"
)

func newPlanTripHandler(anonymous bool) cst.ErrorHandler {
	return cst.ErrorHandler(func(w http.ResponseWriter, r *http.Request) (error, cst.ErrorResponse) {
		tripId, present := mux.Vars(r)["id"]
		if !present {
			return nil, newInvalidIdError()
		}

		var expectDateStr, budgetStr, budgetUnitStr, transportStr string
		var err error
		row := cst.Db.QueryRow("select ExpectedDate, BudgetLimit, BudgetUnit, PreferredTransportMode from ? where TripId = ?", cst.Ev.Var(cst.SqlTripTableVar), tripId)
		if err = row.Scan(&expectDateStr, &budgetStr, &budgetUnitStr, &transportStr); err == sql.ErrNoRows {
			return err, newInvalidIdError()
		}

		var expectDate time.Time
		expectDate, err = time.Parse(cst.DatetimeFormat, expectDateStr)
		if err != nil {
			return err, newServerParseError()
		}

		var budget int
		budget, err = strconv.Atoi(budgetStr)
		if err != nil {
			return err, newServerParseError()
		}

		allowedUnits := utils.NewSet[string]("JPY", "USD")
		allowedTrans := utils.NewSet[string]("train", "bus", "walk")
		if !allowedUnits.Contains(budgetUnitStr) || allowedTrans.Contains(transportStr) {
			return errors.New(""), newServerParseError()
		}

		// Get the points in the trip
		var rows *sql.Rows
		rows, err = cst.Db.Query("select Id, GeoPointId, Name, Lat, Lon from ? where TripId = ?", cst.Ev.Var(cst.SqlPointTableVar), tripId)
		if err != nil {
			return err, newDatabaseQueryError()
		}

		points := map[PointId]Point{}
		for rows.Next() {
			var pid PointId
			var gpid GeoPointId
			var name string
			var lat, lon float64
			err = rows.Scan(&pid, &gpid, &name, &lat, &lon)
			if err != nil {
				return err, newUnknownError()
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
			return rows.Err(), newDatabaseQueryError()
		}

		// Get the point constraints and construct edges
		rows, err = cst.Db.Query("select FirstPointId, SecondPointId, ConstraintType from ? where FirstPointId in ("+strings.Join(pids.Values(), ",")+")", cst.PointConstraintTable)
		if err != nil {
			return err, newDatabaseQueryError()
		}

		for rows.Next() {
			var pid1, pid2 PointId
			var t string
			rows.Scan(&pid1, &pid2, &t)
			p1, ok := points[pid1]
			if !ok {
				return fmt.Errorf("unknown point id %s", pid1), newServerParseError()
			}
			p2, ok := points[pid2]
			if !ok {
				return fmt.Errorf("unknown point id %s", pid1), newServerParseError()
			}
			switch t {
			case "before":
				p1.Next = append(p1.Next, p2)
			case "after":
				p2.Next = append(p2.Next, p1)
			case "first":
				p1.First = true
			case "last":
				p1.Last = true
			default:
				return fmt.Errorf("unknown constraint type %s", t), newServerParseError()
			}
		}

		// Get the candidate orders based on constraints
		var candTrips []PointOrder
		candTrips, err = topologicalSort(utils.GetMapValues[PointId, Point](points), cst.MaxCandidateTrips)

		if err != nil {
			switch err := err.(type) {
			case CycleError:
				var errs []cst.ErrorDescriptor
				for _, ge := range []GraphError(err) {
					var cycle []string
					for _, pid := range []PointId(ge) {
						cycle = append(cycle, points[pid].Name)
					}
					errs = append(errs, cst.ErrorDescriptor{
						Domain:  cst.WebServiceName,
						Reason:  "Cycle found in graph",
						Message: strings.Join(cycle, ","),
					})
				}
				return err, cst.ErrorResponse{
					Code:    http.StatusInternalServerError,
					Message: "Nodes form cycle(s)",
					Errors:  errs,
				}
			case MultiFirstError, MultiLastError, SimulFirstAndLastError, UnknownNodeIdError:
				return err, cst.ErrorResponse{
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
			nps, err = GetNearbyRoutePoints(p, cst.RouteSearchRadius)
			if err != nil {
				return err, newUnknownError()
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
	})
}

func topologicalSort(points []Point, lim int) ([]PointOrder, error) {

	intIds := map[PointId]int{}
	pointIds := map[int]PointId{}

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

	mapback := func(intIds []int) []PointId {
		var ids []PointId
		for _, id := range intIds {
			ids = append(ids, pointIds[id])
		}
		return ids
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

	// BFS
	var unknown []PointId
	indeg := make([]int, len(intIds))
	adj := make([][]int, len(intIds))
	for pid, p := range points {
		for _, next := range p.Next {
			j, ok := intIds[next.Id]
			if !ok {
				unknown = append(unknown, next.Id)
				continue
			}
			indeg[j]++
			adj[intIds[PointId(pid)]] = append(adj[j], intIds[PointId(pid)])
		}
	}

	if len(unknown) > 0 {
		return nil, UnknownNodeIdError(unknown)
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

	// O(N^3) because we want to check for more than one route
	var res []PointOrder
	var dfs func([]int, []int)
	dfs = func(q, cur []int) {
		if len(res) >= lim {
			return
		}
		if len(q) == 0 {
			res = append(res, PointOrder(mapback(cur)))
			return
		}

		var qc []int
		qc = append(qc, q...)
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
			dfs(qc, cur)
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
	dfs(q, cur)

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
				// Found c
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

func GetNearbyRoutePoints(point Point, dist float64) ([]GeoPoint, error) {
	// Get lat lon of bottom left and top right of a square centered at current position and has side = 2*dist
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

	q := fmt.Sprintf("select RouteId, GeoPointId, NodeLat, NodeLon from %s where GeoHashId in (?)", cst.SqlWayTableVar)
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

func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	r := 6378.137e3
	a := math.Sin((lat2 - lat1) / 2)
	b := math.Sin((lon2 - lon1) / 2)
	return 2 * r * math.Asin(math.Sqrt(a*a+math.Cos(lat1)*math.Cos(lat2)*b*b))
}
