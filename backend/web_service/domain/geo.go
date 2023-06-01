package domain

import (
	"errors"
	"fmt"
	"math"
	"time"

	ds "github.com/buihoanganhtuan/tripplanner/backend/web_service/datastructure"
)

const (
	GeohashLen            = 41
	edgeDiffWeight        = 190
	contractionCostWeight = 1
	originEdgesWeight     = 600
	timeEpsilon           = 1e-4
)

type GeoPoint struct {
	Id      GeoPointId     `json:"id"`
	Level   int            `json:"level"`
	Lat     float64        `json:"lat"`
	Lon     float64        `json:"lon"`
	Name    *string        `json:"name,omitempty"`
	Address Address        `json:"address"`
	Tags    []KeyValuePair `json:"tags,omitempty"`
}

type GeoPointId string

type Route struct {
	Id     RouteId        `json:"id"`
	Oneway bool           `json:"oneway"`
	Tags   []KeyValuePair `json:"tags,omitempty"`
}

type RouteId string

type RoutePoint struct {
	GeoPointId    GeoPointId
	RouteId       RouteId
	GeoPointOrder int
}

type GeoEdgeId int64

type GeoEdge struct {
	Id                 GeoEdgeId
	From               GeoPointId
	To                 GeoPointId
	OriginalEdges      int
	Transport          string
	Cost               float64
	TravelTimeFunction []TravelTime
	FixedTravelTime    bool
	LeftChild          *GeoEdgeId
	RightChild         *GeoEdgeId
	MiddleVertex       *GeoPointId
}

type TravelTime struct {
	At    int
	Value int
}

type byImportance struct {
	id         GeoPointId
	importance float64
}

func (v1 byImportance) Less(v2 byImportance) bool {
	return v1.importance < v2.importance
}

func (v byImportance) Id() GeoPointId {
	return v.id
}

type byArrival struct {
	id  GeoPointId
	ttf []TravelTime
}

func (v1 byArrival) Less(v2 byArrival) bool {

}

func (d *Domain) rebuildHierarchies(transport string) error {
	vertices, err := d.GeoRepo.ListTransitNodes()
	if err != nil {
		return err
	}

	// Do local search to get initial ordering
	importPq := ds.NewPriorityQueue[byImportance, GeoPointId]()
	for _, v := range vertices {
		imp, err := d.contract(v.Id, transport, 0, importPq, true)
		if err != nil {
			return err
		}
		importPq.Push(byImportance{v.Id, calcWeight(imp)})
	}

	// importPq now contains initially ordered vertices. Contraction process starts from here
	var level int
	for lazyCount := 0; !importPq.Empty(); level++ {
		fmt.Printf("start processing level %v \n", level)
		then := time.Now()
		lazyThreshold := 1000
		for lazyCount < lazyThreshold {
			curVertex := importPq.Peek()
			imp, err := d.contract(curVertex.id, transport, level, importPq, true)
			if err != nil {
				return err
			}
			importPq.Push(byImportance{curVertex.id, calcWeight(imp)})
			nextVertex := importPq.Peek()
			if curVertex.id == nextVertex.id {
				// current vertex is truly the least important
				break
			}
			lazyCount++
		}

		if lazyCount >= lazyThreshold {
			then := time.Now()
			fmt.Printf("[%v] too many lazy updates. Rebuilding the whole remaining queue... \n", then.Format(time.Stamp))
			newPq := ds.NewPriorityQueue[byImportance, GeoPointId]()
			for !importPq.Empty() {
				v := importPq.Poll()
				imp, err := d.contract(v.id, transport, level, importPq, true)
				if err != nil {
					return err
				}
				newPq.Push(byImportance{v.id, calcWeight(imp)})
			}
			importPq = newPq
			lazyCount = 0
			now := time.Now()
			fmt.Printf("[%v] finish rebuilding queue. Took %v seconds \n", now.Format(time.Stamp), int(now.Sub(then).Seconds()))
		}

		curVertex := importPq.Peek()
		_, err := d.contract(curVertex.id, transport, level, importPq, false)
		if err != nil {
			return err
		}
		now := time.Now()
		fmt.Printf("[%v] finish processing level %v, took %v seconds \n", now.Format(time.Stamp), level, int(time.Now().Sub(then).Seconds()))
	}
	return nil
}

type Importance struct {
	edgeDiff      int
	contractCost  int
	originalEdges int
	numSegments   int
}

// This function performs contraction on a node with the given id. The contraction can be a simulated what-if i.e.,
// it tries to find the _real_ contraction priority of this node given current node ordering and level of contraction
// (number of already contracted nodes) to facilitate lazy update but does not actually update anything. The contraction
// can also be a real contraction, e.g., it actually assigns the current level to the node, add the shortcuts to a persistent
// storage, update the neighbors, etc, which may affect results of subsequent calls to the function.
func (d *Domain) contract(id GeoPointId, transport string, level int, importPq *ds.PriorityQueue[byImportance, GeoPointId], simulated bool) (Importance, error) {
	fEdges, err := d.GeoRepo.EdgesFrom(id, transport)
	if err != nil {
		return Importance{}, err
	}
	rEdges, err := d.GeoRepo.EdgesTo(id, transport)
	if err != nil {
		return Importance{}, err
	}

	var ma distance
	targets := ds.NewSet[GeoPointId]()
	neighbors := ds.NewSet[GeoPointId]()
	for _, fe := range fEdges {
		if level > 0 && !importPq.Exist(fe.To) {
			continue
		}
		neighbors.Add(fe.To)
		targets.Add(fe.To)
		d := distance{fe.Cost, fe.TimeCost}
		if !d.lessOrEqual(ma) {
			ma = d
		}
		break
	}

	var newShortcuts []GeoEdge
	var contractCost, originalEdgeCount int
	for _, re := range rEdges {
		if level > 0 && !importPq.Exist(re.From) {
			continue
		}
		neighbors.Add(re.From)
		src := re.From
		pq := ds.NewPriorityQueue[byDistance, GeoPointId]()
		dist := ds.NewMap[GeoPointId, distance]()
		pq.Push(byDistance{src, distance{0, 0}})
		dist.Put(src, distance{0, 0})
		for tgtCount := 0; !pq.Empty(); contractCost++ {
			v := pq.Poll()
			if !v.distance.lessOrEqual(ma.add(distance{re.Cost, re.TimeCost})) {
				break
			}
			if targets.Contains(v.id) {
				tgtCount++
			}
			if tgtCount == targets.Size() {
				break
			}
			edges, err := d.GeoRepo.EdgesFrom(v.id, transport)
			if err != nil {
				return Importance{}, err
			}
			for _, e := range edges {
				if e.To == id || level > 0 && !importPq.Exist(e.To) {
					// ignore if this neighbor is the current vertex or belong to a lower level, i.e., already contracted
					continue
				}
				newDist := v.distance.add(distance{e.Cost, e.TimeCost})
				if dist.Exist(e.To) && dist.Get(e.To).lessOrEqual(newDist) {
					continue
				}
				dist.Put(e.To, newDist)
				pq.Push(byDistance{e.To, newDist})
			}
		}
		// establish shortcut(s) originating from src
		for _, fe := range fEdges {
			tgt := fe.To
			d := distance{re.Cost + fe.Cost, re.TimeCost + fe.TimeCost}
			if dist.Exist(tgt) && !d.less(dist.Get(tgt)) {
				continue
			}
			newShortcuts = append(newShortcuts, GeoEdge{
				Id:            -1,
				From:          src,
				To:            tgt,
				Transport:     transport,
				Cost:          d.val,
				TimeCost:      d.time,
				OriginalEdges: fe.OriginalEdges + re.OriginalEdges,
				LeftChild:     &fe.Id,
				RightChild:    &re.Id,
				MiddleVertex:  &id,
			})
			originalEdgeCount += fe.OriginalEdges + re.OriginalEdges
		}
	}

	ret := Importance{
		edgeDiff:      len(newShortcuts) - neighbors.Size(),
		contractCost:  contractCost,
		originalEdges: originalEdgeCount,
	}

	if simulated {
		return ret, nil
	}
	d.GeoRepo.PutShortcuts(newShortcuts, transport)
	d.GeoRepo.PutNodeLevel(id, level, transport)
	// update the immediate neighbor
	for _, n := range neighbors.Values() {
		imp, err := d.contract(n, transport, level+1, importPq, true)
		if err != nil {
			return Importance{}, nil
		}
		importPq.Push(byImportance{n, calcWeight(imp)})
	}

	return ret, nil
}

// This function merges travel time functions (TTFs) of src-mid edge and mid-target edge
// to create a new TTF for the src-target shortcut. The input TTFs of the input edges are
// assumed to have FIFO property i.e., t+f(t) < t'+f(t') if t < t'. In other words,
// it assumes that a later departure from a station cannot arrive earlier than a sooner
// departure from the same station. Also, it is assumed that the input TTFs are either
// fixed or periodic with a 1-day period.
func addEdges(edgeIn, edgeOut GeoEdge) GeoEdge {
	const minsPerDay = 1440
	ret := GeoEdge{
		Id:            -1,
		From:          edgeIn.From,
		To:            edgeOut.To,
		OriginalEdges: edgeIn.OriginalEdges + edgeOut.OriginalEdges,
		Transport:     edgeIn.Transport,
		Cost:          edgeIn.Cost + edgeOut.Cost,
		LeftChild:     &edgeIn.Id,
		RightChild:    &edgeOut.Id,
	}
	if edgeIn.FixedTravelTime && edgeOut.FixedTravelTime {
		t0 := edgeIn.TravelTimeFunction[0].Value + edgeOut.TravelTimeFunction[0].Value
		ret.FixedTravelTime = true
		ret.TravelTimeFunction = []TravelTime{TravelTime{Value: t0}}
		return ret
	}
	ret.FixedTravelTime = false
	if edgeOut.FixedTravelTime {
		tmp := make([]TravelTime, len(edgeIn.TravelTimeFunction))
		t0 := edgeOut.TravelTimeFunction[0].Value
		for i, t := range edgeIn.TravelTimeFunction {
			newT := t.Value + t0
			tmp[i] = TravelTime{At: t.At, Value: newT}
		}
		ret.TravelTimeFunction = tmp
		return ret
	}

	if edgeIn.FixedTravelTime {
		tmp := make([]TravelTime, len(edgeOut.TravelTimeFunction))
		t0 := edgeIn.TravelTimeFunction[0].Value
		for i, t := range edgeOut.TravelTimeFunction {
			newT := t.Value + t0
			tmp[i] = TravelTime{At: (t.At - t0 + 1440) % 1440, Value: newT}
		}
		ret.TravelTimeFunction = tmp
		return ret
	}

	ttfIn := edgeIn.TravelTimeFunction
	ttfOut := edgeOut.TravelTimeFunction
	merged := make([]TravelTime, len(ttfIn))

	var j int
	for i := 0; i < len(ttfIn); i++ {
		arr := ttfIn[i].At + ttfIn[i].Value // TODO: transfer time in addition to travel time
		offset := (arr / 1440) * 1440
		for j < len(ttfOut) && ttfOut[j].At+offset < arr {
			j++
		}
		if j == len(ttfOut) {
			offset += 1440
			j = 0
		}
		wait := offset + ttfOut[j].At - arr
		merged[i] = TravelTime{At: ttfIn[i].At, Value: ttfIn[i].Value + wait + ttfOut[j].Value}
	}
	ret.TravelTimeFunction = merged

	return ret
}

func compareTTFs(ttf1, ttf2 []TravelTime) int {
	mi1 := math.MaxInt
	mi2 := math.MaxInt
	for _, t := range ttf1 {
		mi1 = min[int](t.Value, mi1)
	}
	for _, t := range ttf2 {
		mi2 = min[int](t.Value, mi2)
	}
	return mi1 - mi2
}

func calcWeight(imp Importance) float64 {
	return float64(imp.edgeDiff*edgeDiffWeight +
		imp.contractCost*contractionCostWeight +
		imp.originalEdges*originEdgesWeight)
}

// Function to update the shortcuts when an edge changes. Note that this update
// is temporarily only, as we'll be occasionally rebuilding the graph
func (d *Domain) updateEdge(edge GeoEdge, transport string) error {
	edges, err := d.GeoRepo.AncestorEdges(edge.Id, transport)
}

// Internal type used for shortest path finding
type byDistance struct {
	id GeoPointId
	distance
}

func (d distance) less(d2 distance) bool {
	if d.val == d2.val {
		return d.time < d2.time
	}
	return d.val < d2.val
}

func (d distance) equal(d2 distance) bool {
	return d.val == d2.val && d.time == d2.time
}

func (d distance) lessOrEqual(d2 distance) bool {
	if d.val == d2.val {
		return d.time <= d2.time
	}
	return d.val < d2.val
}

func (d distance) add(d2 distance) distance {
	return distance{
		val:  d.val + d2.val,
		time: d.time + d2.time,
	}
}

func (v byDistance) Less(v2 byDistance) bool {
	return v.distance.less(v2.distance)
}

func (v byDistance) Id() GeoPointId {
	return v.id
}

func (d *Domain) shortestPath(src GeoPointId, dst GeoPointId, transport string, time DateTime, budget Cost, duration Duration) ([]GeoEdge, error) {
	// bidirectional dijkstra
	fDist := ds.NewMap[GeoPointId, distance]()
	fPar := ds.NewMap[GeoPointId, GeoEdge]()
	fVis := ds.NewSet[GeoPointId]()
	fPq := ds.NewPriorityQueue[byDistance, GeoPointId]()
	fPq.Push(byDistance{src, distance{0, 0}})

	rDist := ds.NewMap[GeoPointId, distance]()
	rPar := ds.NewMap[GeoPointId, GeoEdge]()
	rVis := ds.NewSet[GeoPointId]()
	rPq := ds.NewPriorityQueue[byDistance, GeoPointId]()
	rPq.Push(byDistance{dst, distance{0, 0}})

	// bidirectional dijkstra
	var sharedVertex *GeoPointId
	minDist := distance{math.MaxFloat64, math.MaxInt}
	for i := 0; !fPq.Empty() || !rPq.Empty(); i++ {
		pq1 := fPq
		d1 := fDist
		v1 := fVis
		p1 := fPar
		d2 := rDist
		v2 := rVis
		if i%2 == 1 {
			pq1 = rPq
			d1 = rDist
			v1 = rVis
			p1 = rPar
			d2 = fDist
			v2 = fVis
		}

		if pq1.Empty() {
			continue
		}
		v := pq1.Poll()
		if !v.distance.less(minDist) {
			break
		}
		if v2.Contains(v.id) {
			if !v.distance.add(d2.Get(v.id)).less(minDist) {
				continue
			}
			sharedVertex = &v.id
			minDist = v.distance.add(d2.Get(v.id))
		}
		v1.Add(v.id)
		var edges []GeoEdge
		var err error
		if i%2 == 0 {
			edges, err = d.GeoRepo.EdgesToUpper(v.id, transport)
		} else {
			edges, err = d.GeoRepo.EdgesFromUpper(v.id, transport)
		}

		if err != nil {
			return nil, err
		}
		for _, e := range edges {
			newDist := v.distance.add(distance{e.Cost, e.TimeCost})
			if v1.Contains(e.To) || d1.Exist(e.To) && !d1.Get(e.To).less(newDist) {
				continue
			}
			d1.Put(e.To, newDist)
			p1.Put(e.To, e)
			pq1.Push(byDistance{e.To, newDist})
		}
	}
	if sharedVertex == nil {
		return nil, nil
	}

	// path
	var path []GeoEdge
	for id := *sharedVertex; id != src; {
		inEdge := fPar.Get(id)
		path = append(path, inEdge)
		id = inEdge.From
	}
	reverse(path)

	for id := *sharedVertex; id != dst; {
		outEdge := rPar.Get(id)
		path = append(path, outEdge)
		id = outEdge.To
	}

	// resolve shortcut
	var unpackedPath []GeoEdge
	for i := 0; i < len(path); i++ {
		unpack, err := d.GeoRepo.ResolveEdge(path[i].Id, transport)
		if err != nil {
			return nil, err
		}
		unpackedPath = append(unpackedPath, unpack...)
	}

	return path, nil

}

// Functions to handle piecewise linear functions. The functions must satisfy following conditions:
//  1. Span the same domain (min time/max time of the two functions must be equal)
//  2. Values must all be positive, times must be all nonnegative
//  3. Has no zero-length or overlapping segments

// a segment represents a half-open interval [start, end)
type segment struct {
	start      float64
	end        float64
	valueStart float64
	valueEnd   float64
}

type piecewiseLinearFunc []segment

// g(t) = t + f(t) can be constructed using this function. It is equivalent to g(t) = h(t) + f(t) where h(t) = t
func addFun(f1, f2 piecewiseLinearFunc, subtract bool) (piecewiseLinearFunc, error) {
	// ensure that two functions have the same domain
	if f1[0].start != f2[0].start || f1[len(f1)-1].end != f2[len(f2)-1].end {
		return nil, errors.New("two functions must have the same domain")
	}
	var f piecewiseLinearFunc
	var i, j int
	var sgn float64 = 1
	if subtract {
		sgn = -1
	}
	for i < len(f1) && j < len(f2) {
		var s1, s2 segment = f1[i], f2[j]
		a1 := (s1.valueStart - s1.valueEnd) / (s1.start - s1.end)
		a2 := (s2.valueStart - s2.valueEnd) / (s2.start - s2.end)
		b1 := s1.valueStart - a1*s1.start
		b2 := s2.valueStart - a2*s2.start
		l := max[float64](s1.start, s2.start)
		r := min[float64](s1.end, s2.end)
		f = append(f, segment{start: l, end: r, valueStart: (a1+a2)*l + sgn*(b1+b2), valueEnd: (a1+a2)*r + sgn*(b1+b2)})
		if s1.end < s2.end {
			i++
		}
		if s1.end > s2.end {
			j++
		}
		if s1.end == s2.end {
			i++
			j++
		}
	}
	return f, nil
}

func minimumFun(f1, f2 piecewiseLinearFunc, checkOnly bool) (piecewiseLinearFunc, bool, error) {
	// ensure that two functions have the same domain
	if f1[0].start != f2[0].start || f1[len(f1)-1].end != f2[len(f2)-1].end {
		return nil, false, errors.New("two functions must have the same domain")
	}
	var i, j int
	var f piecewiseLinearFunc
	var less bool = true

	for i < len(f1) && j < len(f2) {
		var s1, s2 segment = f1[i], f2[j]
		a1 := (s1.valueStart - s1.valueEnd) / (s1.start - s1.end)
		a2 := (s2.valueStart - s2.valueEnd) / (s2.start - s2.end)
		b1 := s1.valueStart - a1*s1.start
		b2 := s2.valueStart - a2*s2.start
		l := max[float64](s1.start, s2.start)
		r := min[float64](s1.end, s2.end)
		t := (b1 - b2) / (a2 - a1)
		if t <= l+timeEpsilon || t >= r-timeEpsilon {
			// no intersection
			f = append(f, segment{start: l, end: r, valueStart: min[float64](a1*l+b1, a2*l+b2), valueEnd: min[float64](a1*r+b1, a2*r+b2)})
		} else {
			less = false
			if checkOnly {
				break
			}
			f = append(f, segment{start: l, end: t, valueStart: min[float64](a1*l+b1, a2*l+b2), valueEnd: a1*t + b1}, segment{start: t, end: r, valueStart: a1*t + b1, valueEnd: min[float64](a1*r+b1, a2*r+b2)})
		}
		if s1.end < s2.end {
			i++
		}
		if s1.end > s2.end {
			j++
		}
		if s1.end == s2.end {
			i++
			j++
		}
	}
	return f, less, nil

}

func composeFun(f1, f2 piecewiseLinearFunc) (piecewiseLinearFunc, error) {
	// truncate range (y) of f2 so that it is wholly contained by domain (x) of f2
	mi := f2[0].start + timeEpsilon
	ma := f2[len(f2)-1].end - timeEpsilon
	var _f2 piecewiseLinearFunc
	for _, s := range f2 {
		a := (s.valueEnd - s.valueStart) / (s.end - s.start)
		b := s.valueStart - a*s.start
		if a == 0 {
			return nil, errors.New("range of a segment from f2 is zero-length")
		}

		tmin := (mi - b) / a
		tmax := (ma - b) / a
		if tmin > tmax {
			tmin, tmax = tmax, tmin
		}
		tmin = max[float64](tmin, s.start)
		tmax = min[float64](tmax, s.end)
		if tmin > tmax {
			continue
		}
		_f2 = append(_f2, segment{start: tmin, end: tmax, valueStart: a*tmin + b, valueEnd: a*tmax + b})
	}

	var f piecewiseLinearFunc
	for _, s2 := range _f2 {
		a2 := (s2.valueEnd - s2.valueStart) / (s2.end - s2.start)
		b2 := s2.valueEnd - a2*s2.end
		if s2.valueStart < s2.valueEnd {
			i := lastLeq[segment, float64](f1, s2.valueStart, func(seg segment, val float64) bool {
				return seg.start <= val
			})
			j := lastLeq[segment, float64](f1, s2.valueEnd, func(seg segment, val float64) bool {
				return seg.end <= val
			})
			for i <= j {
				s1 := f1[i]
				start := max[float64](s2.valueStart, s1.start)
				end := min[float64](s2.valueEnd, s1.end)
				a1 := (s1.valueEnd - s1.valueStart) / (s1.end - s1.start)
				b1 := s1.valueEnd - a1*s1.end
				f = append(f, segment{
					start:      (start - b2) / a2,
					end:        (end - b2) / a2,
					valueStart: a1*start + b1,
					valueEnd:   a1*end + b1,
				})
				i++
			}
		} else {
			i := lastLeq[segment, float64](f1, s2.valueStart, func(seg segment, val float64) bool {
				return seg.start <= val
			})
			j := lastLeq[segment, float64](f1, s2.valueEnd, func(seg segment, val float64) bool {
				return seg.start <= val
			})
			for i >= j {
				s1 := f1[i]
				start := min[float64](s2.valueStart, s1.end)
				end := max[float64](s2.valueEnd, s1.start)
				a1 := (s1.valueEnd - s1.valueStart) / (s1.end - s1.start)
				b1 := s1.valueEnd - a1*s1.end
				f = append(f, segment{
					start:      (start - b2) / a2,
					end:        (end - b2) / a2,
					valueStart: a1*start + b1,
					valueEnd:   a1*end + b1,
				})
				i--
			}
		}
	}
	return f, nil
}

func minFun(f piecewiseLinearFunc) float64 {
	v := f[0].valueStart
	for i := 1; i < len(f); i++ {
		v = min[float64](v, f[i].valueStart)
	}
	return v
}

func reverse[T any](arr []T) {
	var i int
	var j int = len(arr)
	for i < j {
		arr[i], arr[j] = arr[j], arr[i]
		i++
		j--
	}
}

func min[T ~int | ~int32 | ~int64 | float32 | float64 | string](arr ...T) T {
	ret := arr[0]
	for i := 1; i < len(arr); i++ {
		if ret > arr[i] {
			ret = arr[i]
		}
	}
	return ret
}

func max[T ~int | ~int32 | ~int64 | float32 | float64 | string](arr ...T) T {
	ret := arr[0]
	for i := 1; i < len(arr); i++ {
		if ret < arr[i] {
			ret = arr[i]
		}
	}
	return ret
}

func firstGreater[T any, V ~int | ~float64](arr []T, target V, gfun func(T, V) bool) int {
	var l, r int = 0, len(arr) - 1
	for l < r {
		m := (l + r) / 2
		if gfun(arr[m], target) {
			r = m
		} else {
			l = m + 1
		}
	}
	return l
}

func lastLeq[T any, V ~int | ~float64](arr []T, target V, leqFun func(T, V) bool) int {
	var l, r int = 0, len(arr) - 1
	for l < r {
		m := (l + r + 1) / 2
		if leqFun(arr[m], target) {
			l = m
		} else {
			r = m - 1
		}
	}
	return l
}
