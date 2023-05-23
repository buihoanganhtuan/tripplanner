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
	Id            GeoEdgeId
	EndPoint1     GeoPointId
	EndPoint2     GeoPointId
	OriginalEdges int
	Transport     string
	Cost          float64
	LeftChild     *GeoEdgeId
	RightChild    *GeoEdgeId
	MiddleVertex  *GeoPointId
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

func (d *Domain) contract(id GeoPointId, transport string, level int, importPq *ds.PriorityQueue[vertexByImportance, GeoPointId], simulated bool) (Importance, error) {
	fEdges, err := d.GeoRepo.EdgesFrom(id, transport)
	if err != nil {
		return Importance{}, err
	}
	rEdges, err := d.GeoRepo.EdgesTo(id, transport)
	if err != nil {
		return Importance{}, err
	}

	var ma float64
	targets := ds.NewSet[GeoPointId]()
	neighbors := ds.NewSet[GeoPointId]()
	for _, fe := range fEdges {
		if !importPq.Exist(fe.EndPoint2) {
			continue
		}
		neighbors.Add(fe.EndPoint2)
		targets.Add(fe.EndPoint2)
		ma = math.Max(ma, fe.Cost)
		break
	}

	var newShortcuts []GeoEdge
	var contractCost, orgEdgeCount int
	for _, re := range rEdges {
		if !importPq.Exist(re.EndPoint1) {
			continue
		}
		neighbors.Add(re.EndPoint1)
		src := re.EndPoint1
		pq := ds.NewPriorityQueue[vertexByDistance, GeoPointId]()
		dist := ds.NewMap[GeoPointId, float64]()
		pq.Push(vertexByDistance{src, 0})
		dist.Put(src, 0)
		for tgtCount := 0; !pq.Empty(); contractCost++ {
			v := pq.Poll()
			if v.dist > ma+re.Cost {
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
				if e.EndPoint2 == id || !importPq.Exist(e.EndPoint2) {
					// ignore if this neighbor is the current vertex or belong to a lower level, i.e., already contracted
					continue
				}
				newDist := v.dist + e.Cost
				if dist.Exist(e.EndPoint2) && dist.Get(e.EndPoint2) <= newDist {
					continue
				}
				dist.Put(e.EndPoint2, newDist)
				pq.Push(vertexByDistance{e.EndPoint2, newDist})
			}
		}
		// establish shortcut(s) originating from src
		for _, fe := range fEdges {
			tgt := fe.EndPoint2
			if dist.Exist(tgt) && re.Cost+fe.Cost >= dist.Get(tgt) {
				continue
			}
			newShortcuts = append(newShortcuts, GeoEdge{
				Id:            -1,
				EndPoint1:     src,
				EndPoint2:     tgt,
				Transport:     transport,
				Cost:          re.Cost + fe.Cost,
				OriginalEdges: fe.OriginalEdges + re.OriginalEdges,
				LeftChild:     &fe.Id,
				RightChild:    &re.Id,
				MiddleVertex:  &id,
			})
			orgEdgeCount += fe.OriginalEdges + re.OriginalEdges
		}
	}

	ret := Importance{
		edgeDiff:      len(newShortcuts) - neighbors.Size(),
		contractCost:  contractCost,
		originalEdges: orgEdgeCount,
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
	id   GeoPointId
	dist float64
}

func (v vertexByDistance) Less(v2 vertexByDistance) bool {
	return v.dist < v2.dist
}

func (v vertexByDistance) Id() GeoPointId {
	return v.id
}

func (d *Domain) shortestPath(src GeoPointId, dst GeoPointId, transport string) ([][]GeoEdge, error) {
	// bidirectional dijkstra
	fDist := ds.NewMap[GeoPointId, float64]()
	fPar := ds.NewMap[GeoPointId, GeoEdge]()
	fVis := ds.NewSet[GeoPointId]()
	fPq := ds.NewPriorityQueue[vertexByDistance, GeoPointId]()
	fPq.Push(vertexByDistance{src, 0})

	rDist := ds.NewMap[GeoPointId, float64]()
	rPar := ds.NewMap[GeoPointId, GeoEdge]()
	rVis := ds.NewSet[GeoPointId]()
	rPq := ds.NewPriorityQueue[vertexByDistance, GeoPointId]()
	rPq.Push(vertexByDistance{dst, 0})

	// bidirectional dijkstra
	var sharedVertices []GeoPointId
	minDist := math.MaxFloat64
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
		if v.dist > minDist {
			continue
		}
		if v2.Contains(v.id) {
			if v.dist+d2.Get(v.id) < minDist {
				sharedVertices = nil
				minDist = v.dist + d2.Get(v.id)
			}
			if v.dist+d2.Get(v.id) == minDist {
				sharedVertices = append(sharedVertices, v.id)
			}
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
			newDist := v.dist + e.Cost
			if v1.Contains(e.EndPoint2) || d1.Exist(e.EndPoint2) && d1.Get(e.EndPoint2) >= newDist {
				continue
			}
			d1.Put(e.EndPoint2, newDist)
			p1.Put(e.EndPoint2, e)
			pq1.Push(vertexByDistance{e.EndPoint2, newDist})
		}
	}
	if len(sharedVertices) == 0 {
		return nil, nil
	}

	// path
	var paths [][]GeoEdge
	for _, v := range sharedVertices {
		var path []GeoEdge
		for id := v; id != src; {
			inEdge := fPar.Get(id)
			path = append(path, inEdge)
			id = inEdge.EndPoint1
		}
		reverse(path)

		for id := v; id != dst; {
			outEdge := rPar.Get(id)
			path = append(path, outEdge)
			id = outEdge.EndPoint2
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

		paths = append(paths, path)
	}

	return paths, nil

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
