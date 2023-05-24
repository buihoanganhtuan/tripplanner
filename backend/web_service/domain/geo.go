package domain

import (
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
	LeftChild          *GeoEdgeId
	RightChild         *GeoEdgeId
	MiddleVertex       *GeoPointId
}

type TravelTime struct {
	Enter    DateTime
	TimeCost int
}

type vertexByImportance struct {
	id         GeoPointId
	importance float64
}

func (v vertexByImportance) Less(v1 vertexByImportance) bool {
	return v.importance < v1.importance
}

func (v vertexByImportance) Id() GeoPointId {
	return v.id
}

func (d *Domain) rebuildHierarchies(transport string) error {
	vertices, err := d.GeoRepo.ListTransitNodes()
	if err != nil {
		return err
	}

	// Do local search to get initial ordering
	importPq := ds.NewPriorityQueue[vertexByImportance, GeoPointId]()
	for _, v := range vertices {
		imp, err := d.contract(v.Id, transport, 0, importPq, true)
		if err != nil {
			return err
		}
		importPq.Push(vertexByImportance{v.Id, calcWeight(imp)})
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
			importPq.Push(vertexByImportance{curVertex.id, calcWeight(imp)})
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
			newPq := ds.NewPriorityQueue[vertexByImportance, GeoPointId]()
			for !importPq.Empty() {
				v := importPq.Poll()
				imp, err := d.contract(v.id, transport, level, importPq, true)
				if err != nil {
					return err
				}
				newPq.Push(vertexByImportance{v.id, calcWeight(imp)})
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
}

// This function performs contraction on a node with the given id. The contraction can be a simulated what-if i.e.,
// it tries to find the _real_ contraction priority of this node given current node ordering and level of contraction
// (number of already contracted nodes) to facilitate lazy update but does not actually update anything. The contraction
// can also be a real contraction, e.g., it actually assigns the current level to the node, add the shortcuts to a persistent
// storage, update the neighbors, etc, which may affect results of subsequent calls to the function.
func (d *Domain) contract(id GeoPointId, transport string, level int, importPq *ds.PriorityQueue[vertexByImportance, GeoPointId], simulated bool) (Importance, error) {
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
		pq := ds.NewPriorityQueue[vertexByDistance, GeoPointId]()
		dist := ds.NewMap[GeoPointId, distance]()
		pq.Push(vertexByDistance{src, distance{0, 0}})
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
				pq.Push(vertexByDistance{e.To, newDist})
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
		importPq.Push(vertexByImportance{n, calcWeight(imp)})
	}

	return ret, nil
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
type vertexByDistance struct {
	id GeoPointId
	distance
}

type distance struct {
	val  float64
	time int
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

func (v vertexByDistance) Less(v2 vertexByDistance) bool {
	return v.distance.less(v2.distance)
}

func (v vertexByDistance) Id() GeoPointId {
	return v.id
}

func (d *Domain) shortestPath(src GeoPointId, dst GeoPointId, transport string) ([]GeoEdge, error) {
	// bidirectional dijkstra
	fDist := ds.NewMap[GeoPointId, distance]()
	fPar := ds.NewMap[GeoPointId, GeoEdge]()
	fVis := ds.NewSet[GeoPointId]()
	fPq := ds.NewPriorityQueue[vertexByDistance, GeoPointId]()
	fPq.Push(vertexByDistance{src, distance{0, 0}})

	rDist := ds.NewMap[GeoPointId, distance]()
	rPar := ds.NewMap[GeoPointId, GeoEdge]()
	rVis := ds.NewSet[GeoPointId]()
	rPq := ds.NewPriorityQueue[vertexByDistance, GeoPointId]()
	rPq.Push(vertexByDistance{dst, distance{0, 0}})

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
			pq1.Push(vertexByDistance{e.To, newDist})
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

func reverse[T any](arr []T) {
	var i int
	var j int = len(arr)
	for i < j {
		arr[i], arr[j] = arr[j], arr[i]
		i++
		j--
	}
}
