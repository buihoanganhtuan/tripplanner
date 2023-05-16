package domain

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/buihoanganhtuan/tripplanner/backend/web_service/datastructure"
)

var types = datastructure.NewDefaultSet[string]("anon", "reg")
var transport = datastructure.NewDefaultSet[string]("train", "bus", "walk")
var mUnit = datastructure.NewDefaultSet[string]("usd", "jpy")
var dUnit = datastructure.NewDefaultSet[string]("hour", "min")

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

type TripId string

// internal types, not exposed
type graphError []PointId
type cycleError []graphError
type unknownIdError graphError
type pointOrder []PointId
type cycle []int
type denormPoint struct {
	Point
	GeoPoint
}

func (ge graphError) Error() string {
	var pids []string
	for _, pid := range ge {
		pids = append(pids, string(pid))
	}
	return strings.Join(pids, ",")
}

func (un unknownIdError) Error() string {
	return graphError(un).Error()
}

func (ce cycleError) Error() string {
	ges := []graphError(ce)
	var sb strings.Builder
	for _, ge := range ges {
		sb.WriteString(ge.Error())
		sb.WriteString("\\n")
	}
	return sb.String()
}

func (d *Domain) PlanTrip(id TripId) (Trip, error) {
	transId, err := d.repo.CreateTransaction()
	if err != nil {
		return Trip{}, err
	}
	defer d.repo.RollbackTransaction(transId)

	trip, err := d.repo.GetTrip(id, transId)
	if err = validateTrip(trip); err != nil {
		return Trip{}, err
	}

	points, err := d.repo.PointsWithTrip(id)
	if err != nil {
		return Trip{}, err
	}

	if err = validatePoints(points); err != nil {
		return Trip{}, err
	}

	var tripCands []pointOrder
	if trip.Type == "anon" {
		tripCands, err = topologicalSort(points, *trip.DateExpected, 3)
	} else {
		tripCands, err = topologicalSort(points, *trip.DateExpected, 10)
	}

	var gpids []GeoPointId
	for _, p := range points {
		gpids = append(gpids, p.GeoPointId)
	}

	geopoints, err := d.repo.GeoPoints(gpids)
	if err != nil {
		return Trip{}, err
	}

	idx := datastructure.NewMap[PointId, int]()
	for i := 0; i < len(points); i++ {
		idx.Put(points[i].Id, i)
	}

	for _, tripCand := range tripCands {
		var plan []Path
		for i := 0; i < len(tripCand)-1; i++ {
			j := idx.Get(tripCand[i])
			k := idx.Get(tripCand[i+1])
			path, can, err := d.findPaths(
				denormPoint{Point: points[j], GeoPoint: geopoints[j]},
				denormPoint{Point: points[k], GeoPoint: geopoints[k]},
				trip.PreferredMode)
			if err != nil {
				return Trip{}, err
			}
			if !can {
				plan = nil
				break
			}
			plan = append(plan, path)
		}
	}
}

func (d *Domain) findPaths(p1 denormPoint, p2 denormPoint, transport string) (Path, bool, error) {

}

// This function finds the geo points whose distance
// to the input point is not more than dist
func (d *Domain) getNearbyPoints(id GeoPointId, dist float64) ([]GeoPoint, error) {
	geoPoint, err := d.repo.GeoPoint(id)
	if err != nil {
		return nil, err
	}
	lat := geoPoint.Lat
	lon := geoPoint.Lon

	latBits := GeohashLen / 2
	lonBits := GeohashLen/2 + GeohashLen%2
	numLats := int64(1) << int64(latBits)
	numLons := int64(1) << int64(lonBits)
	dlat := float64(180) / float64(numLats)
	dlon := float64(360) / float64(numLons)

	var lo, hi int
	hi = int(lat / dlat)
	for lo < hi {
		mid := (lo + hi + 1) / 2
		mlat := float64(-90) + float64(mid)*dlat
		if haversine(lat, lon, mlat, lon) >= dist {
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	blat := int64(lo)

	lo = int(lat/dlat) + 1
	hi = 1<<latBits - 1
	for lo < hi {
		mid := (lo + hi) / 2
		mlat := float64(-90) + float64(mid)*dlat
		if haversine(lat, lon, mlat, lon) >= dist {
			hi = mid
		} else {
			lo = mid + 1
		}
	}
	tlat := int64(lo)

	lo = 0
	hi = int(lon / dlon)
	for lo < hi {
		mid := (lo + hi + 1) / 2
		mlon := float64(-180) + float64(mid)*dlon
		if haversine(lat, lon, lat, mlon) >= dist {
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	llon := int64(lo)

	lo = int(lon/dlon) + 1
	hi = 1<<lonBits - 1
	for lo < hi {
		mid := (lo + hi) / 2
		mlon := float64(-180) + float64(mid)*dlon
		if haversine(lat, lon, lat, mlon) >= dist {
			hi = mid
		} else {
			lo = mid + 1
		}
	}
	rlon := int64(lo)

	var geoHashes []GeoHashId
	for j := blat; j <= tlat; j++ {
		for k := llon; k <= rlon; k++ {
			geoHashes = append(geoHashes, GeoHashId(strconv.FormatInt(j+k<<latBits, 10)))
		}
	}

	pp, err := d.repo.GeoPointsWithHashes(geoHashes)
	if err != nil {
		return nil, err
	}

	var tmp []GeoPoint
	for _, p := range pp {
		if haversine(lat, lon, p.Lat, p.Lon) > dist {
			continue
		}
		if err = p.validate(); err != nil {
			return nil, err
		}
		tmp = append(tmp, p)
	}

	return tmp, nil
}

func validateTrip(t Trip) error {
	if !types.Contains(t.Type) {
		return errors.New("unknown trip type")
	}
	if t.Type != "anon" && t.UserId == "" {
		return errors.New("non-anonymous trips must belong to a user")
	}
	if !mUnit.Contains(t.Budget.Unit) {
		return errors.New("invalid money unit")
	}
	if !transport.Contains(t.PreferredMode) {
		return errors.New("invalid transport mode")
	}
	if t.DateExpected == nil {
		return errors.New("trip must have an expected date")
	}
	return nil
}

func validatePoints(pp []Point) error {
	pids := datastructure.NewSet[PointId]()
	for _, p := range pp {
		pids.Add(p.Id)
	}

	f := func(pid PointId) string {
		return string(pid)
	}

	for _, p := range pp {
		if p.First && p.Last {
			return errors.New(fmt.Sprintf("point %v: cannot be first and last simultaneously", p.Id))
		}
		if !dUnit.Contains(p.Duration.Unit) {
			return errors.New(fmt.Sprintf("point %v: unknown duration unit", p.Id))
		}
		bf := datastructure.NewDefaultSet[PointId](p.Before.Points...)
		af := datastructure.NewDefaultSet[PointId](p.After.Points...)
		common := bf.Intersection(af)
		if common.Size() != 0 {
			return errors.New(fmt.Sprintf("point %v: both before and after point(s) %v", p.Id, common.ToString(f, ",")))
		}
		ubf := bf.Difference(pids)
		uaf := af.Difference(pids)
		if ubf.Size() > 0 {
			return errors.New(fmt.Sprintf("point %v: unknown `before` point(s) %v", p.Id, ubf.ToString(f, ",")))
		}
		if uaf.Size() > 0 {
			return errors.New(fmt.Sprintf("point %v: unknown `after` point(s) %v", p.Id, uaf.ToString(f, ",")))
		}
	}

	return nil
}

/*
Extract possible solutions to a certain DAG ordering. As the number of solutions can be
quite large, we terminate the search when the number of results found thus far exceed lim
*/

func topologicalSort(points []Point, start DateTime, lim int) ([]pointOrder, error) {
	intIds := datastructure.NewMap[PointId, int]()
	pointIds := datastructure.NewMap[int, PointId]()

	mapback := func(intIds []int) []PointId {
		var ids []PointId
		for _, id := range intIds {
			ids = append(ids, pointIds.Get(id))
		}
		return ids
	}

	// convert PointId (string) to integer id
	for i, p := range points {
		intIds.Put(p.Id, i)
		pointIds.Put(i, p.Id)
	}

	// construct the directed edges, prepare the in-degree count for each node
	indeg := make([]int, intIds.Size())
	adj := make([][]int, intIds.Size())
	for i, p := range points {
		for _, next := range p.Before.Points {
			j := intIds.Get(next)
			indeg[j]++
			adj[i] = append(adj[j], i)
		}
	}

	// Check for cycle
	cycles := findCycles(indeg, adj)
	if cycles != nil {
		var ce []graphError
		for _, c := range cycles {
			ce = append(ce, mapback(c))
		}
		return nil, cycleError(ce)
	}

	sortFn := func(i, j int) bool {
		d1 := points[i].Arrival
		d2 := points[j].Arrival
		if d1 != nil && d2 != nil {
			t1 := points[i].Arrival.Before
			t2 := points[j].Arrival.Before
			return t1.before(t2)
		}
		if d1 == nil && d2 == nil || d1 == nil && d2 != nil {
			return false
		}
		return true
	}

	var res []pointOrder
	var dfs func([]int, []int, DateTime)
	dfs = func(q, cur []int, t DateTime) {
		if len(res) >= lim {
			return
		}
		if len(q) == 0 {
			res = append(res, pointOrder(mapback(cur)))
			return
		}

		var qc []int
		qc = append(qc, q...)
		sort.SliceStable(qc, sortFn) // prioritize points with deadline first

		if points[qc[0]].Arrival != nil && t.after(points[qc[0]].Arrival.Before) {
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

			dfs(qc, cur, t.add(points[qc[i]].Duration))

			for _, j := range adj[tmp] {
				indeg[j]++
			}
			qc = qc[:len(qc)-added]
			qc = append(qc, qc[i])
			qc[i] = tmp
			cur = cur[:len(cur)-1]
		}
	}

	// prepare first nodes for the backtrack process
	var q, cur []int
	for i, d := range indeg {
		if d == 0 {
			q = append(q, i)
		}
	}

	dfs(q, cur, start)

	return res, nil
}

func findCycles(indeg []int, adj [][]int) []cycle {
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

	var res []cycle

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

func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	r := 6378.137e3
	a := math.Sin((lat2 - lat1) / 2)
	b := math.Sin((lon2 - lon1) / 2)
	return 2 * r * math.Asin(math.Sqrt(a*a+math.Cos(lat1)*math.Cos(lat2)*b*b))
}
