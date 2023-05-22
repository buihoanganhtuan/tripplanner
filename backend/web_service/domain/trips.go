package domain

import (
	"errors"
	"fmt"
	"math"
	"sort"
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
	transId, err := d.Repo.CreateTransaction()
	if err != nil {
		return Trip{}, err
	}
	defer d.Repo.RollbackTransaction(transId)

	trip, err := d.Repo.GetTrip(id, transId)
	if err = validateTrip(trip); err != nil {
		return Trip{}, err
	}

	points, err := d.Repo.PointsWithTrip(id)
	if err != nil {
		return Trip{}, err
	}

	if err = validatePoints(points); err != nil {
		return Trip{}, err
	}

	var tripCands []pointOrder
	if trip.Type == "anon" {
		tripCands, err = d.topologicalSort(points, *trip.DateExpected, 3)
	} else {
		tripCands, err = d.topologicalSort(points, *trip.DateExpected, 10)
	}

	var gpids []GeoPointId
	for _, p := range points {
		gpids = append(gpids, p.GeoPointId)
	}

	geopoints, err := d.GeoRepo.GeoPoints(gpids)
	if err != nil {
		return Trip{}, err
	}

	idx := datastructure.NewMap[PointId, int]()
	for i := 0; i < len(points); i++ {
		idx.Put(points[i].Id, i)
	}

	for _, tripCand := range tripCands {
		var plan []Path
		r := 2000.
		if trip.Type == "anon" {
			r = 500.
		}
		for i := 0; i < len(tripCand)-1; i++ {
			j := idx.Get(tripCand[i])
			k := idx.Get(tripCand[i+1])
			path, can, err := d.findPaths(
				denormPoint{Point: points[j], GeoPoint: geopoints[j]},
				denormPoint{Point: points[k], GeoPoint: geopoints[k]},
				r, trip.PreferredMode)
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

func (d *Domain) findPaths(src denormPoint, dst denormPoint, dist float64, transport string) (Path, bool, error) {
	gpp1, err := d.GeoRepo.TransitNodes(src.GeoPoint.Id, dist)
	if err != nil {
		return Path{}, false, err
	}
	gpp2, err := d.GeoRepo.TransitNodes(dst.GeoPoint.Id, dist)
	if err != nil {
		return Path{}, false, err
	}

}

// This function finds the geo points whose distance
// to the input point is not more than dist

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

func (d *Domain) topologicalSort(points []Point, start DateTime, lim int) ([]pointOrder, error) {
	intIds := datastructure.NewMap[PointId, int]()
	pointIds := datastructure.NewMap[int, PointId]()

	mapback := func(intIds []int) []PointId {
		var ids []PointId
		for _, id := range intIds {
			ids = append(ids, pointIds.Get(id))
		}
		return ids
	}

	// convert PointId (string) to integer id. Also, get the geoId list for later use
	var geoIds []GeoPointId
	for i, p := range points {
		intIds.Put(p.Id, i)
		pointIds.Put(i, p.Id)
		geoIds = append(geoIds, p.GeoPointId)
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

	type OrderedPoint struct {
		id         PointId
		arrival    *PointArrivalConstraint
		duration   Duration
		lat        float64
		lon        float64
		distToPrev float64
	}

	var arPoints []OrderedPoint
	geoPoints, err := d.GeoRepo.GeoPoints(geoIds)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(points); i++ {
		arPoints = append(arPoints, OrderedPoint{
			id:       points[i].Id,
			arrival:  points[i].Arrival,
			duration: points[i].Duration,
			lat:      geoPoints[i].Lat,
			lon:      geoPoints[i].Lon,
		})
	}

	less := func(i, j int) bool {
		d1 := arPoints[i].arrival
		d2 := arPoints[j].arrival
		if d1 != nil && d2 != nil {
			t1 := arPoints[i].arrival.Before
			t2 := arPoints[j].arrival.Before
			return t1.before(t2)
		}
		if d1 == nil && d2 == nil {
			return arPoints[i].distToPrev < arPoints[j].distToPrev
		}
		return d2 == nil
	}

	var res []pointOrder
	var dfs func([]int, []int, DateTime)
	dfs = func(queue, cur []int, t DateTime) {
		if len(res) >= lim {
			return
		}
		if len(queue) == 0 {
			res = append(res, pointOrder(mapback(cur)))
			return
		}

		var queueCopy []int
		var orgDist []float64
		defer func() {
			for i := 0; i < len(queueCopy); i++ {
				arPoints[queueCopy[i]].distToPrev = orgDist[i]
			}
		}()
		queueCopy = append(queueCopy, queue...)
		for i := 0; len(cur) > 0 && i < len(queueCopy); i++ {
			orgDist = append(orgDist, arPoints[queueCopy[i]].distToPrev)
			last := cur[len(cur)-1]
			arPoints[queueCopy[i]].distToPrev = haversine(arPoints[last].lat, arPoints[last].lon, arPoints[queueCopy[i]].lat, arPoints[queueCopy[i]].lon)
		}

		sort.SliceStable(queueCopy, less) // prioritize points with deadline first

		if points[queueCopy[0]].Arrival != nil && t.after(points[queueCopy[0]].Arrival.Before) {
			return
		}

		// backtrack
		for i := 0; i < len(queueCopy); i++ {
			tmp := queueCopy[i]
			cur = append(cur, tmp)
			queueCopy[i] = queueCopy[len(queueCopy)-1]
			queueCopy = queueCopy[:len(queueCopy)-1]
			var added int
			for _, j := range adj[tmp] {
				indeg[j]--
				if indeg[j] == 0 {
					added++
					queueCopy = append(queueCopy, j)
				}
			}

			dfs(queueCopy, cur, t.add(points[queueCopy[i]].Duration))

			for _, j := range adj[tmp] {
				indeg[j]++
			}
			queueCopy = queueCopy[:len(queueCopy)-added]
			queueCopy = append(queueCopy, queueCopy[i])
			queueCopy[i] = tmp
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
