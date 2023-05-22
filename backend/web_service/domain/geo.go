package domain

import (
	"math"

	ds "github.com/buihoanganhtuan/tripplanner/backend/web_service/datastructure"
)

const (
	GeohashLen            = 41
	edgeDiffWeight        = 190
	contractionCostWeight = 1
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
	Level         int
	EndPoint1     GeoPointId
	EndPoint2     GeoPointId
	OriginalEdges int
	Transport     string
	Cost          float64
	LeftChild     GeoEdgeId
	RightChild    GeoEdgeId
	Parent        GeoEdgeId
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

func (d *Domain) importanceUpdate(transport string) error {
	vertices, err := d.GeoRepo.ListTransitNodes()
	if err != nil {
		return err
	}

	// Do local search to get initial ordering
	importPq := ds.NewPriorityQueue[vertexByImportance, GeoPointId]()
	for _, curVertex := range vertices {

	}

	// importPq now contains initially ordered vertices. Contraction starts from here
	for !importPq.Empty() {
		curVertex := importPq.Peek()
		// lazy update

	}

}

type Importance struct {
	edgeDiff      int
	contractCost  int
	originalEdges int
}

func (d *Domain) contract(id GeoPointId, transport string, level int, simulated bool) (Importance, error) {
	fEdges, err := d.GeoRepo.EdgesToUpper(id, transport)
	if err != nil {
		return Importance{}, err
	}
	rEdges, err := d.GeoRepo.EdgesFromUpper(id, transport)
	if err != nil {
		return Importance{}, err
	}

	var ma float64
	for _, e := range fEdges {
		ma = math.Max(ma, e.Cost)
		break
	}

	var contrCost, edgeDiff int
	var newShortcuts []GeoEdge
	for _, re := range rEdges {
		src := re.EndPoint1
		pq := ds.NewPriorityQueue[vertexByDistance, GeoPointId]()
		dist := ds.NewMap[GeoPointId, float64]()
		pq.Push(vertexByDistance{
			id:   src,
			dist: 0})
		for !pq.Empty() {
			v := pq.Poll()
			// optimization 1: early termination
			if v.dist > ma+re.Cost {
				break
			}
			dist.Put(v.id, v.dist)
			contrCost++
			edges, err := d.GeoRepo.EdgesToUpper(v.id, transport)
			if err != nil {
				return Importance{}, err
			}
			for _, e := range edges {
				if e.EndPoint2 == id {
					// ignore current node
					continue
				}
				newDist := v.dist + e.Cost
				if dist.Exist(e.EndPoint2) && dist.Get(e.EndPoint2) <= newDist {
					continue
				}
				dist.Put(e.EndPoint2, newDist)
				pq.Push(vertexByDistance{
					id:   e.EndPoint2,
					dist: newDist})
			}
		}
		// find the target of the shortcut from src
		for _, fe := range fEdges {
			tgt := fe.EndPoint2
			if dist.Exist(tgt) && re.Cost+fe.Cost >= dist.Get(tgt) {
				continue
			}
			edgeDiff++
			if simulated {
				continue
			}
			sid, err := d.GeoRepo.NewEdgeId(transport)
			if err != nil {
				return Importance{}, err
			}
			newShortcuts = append(newShortcuts, GeoEdge{
				Id:            sid,
				Level:         level,
				EndPoint1:     src,
				EndPoint2:     tgt,
				Transport:     transport,
				Cost:          re.Cost + fe.Cost,
				OriginalEdges: fe.OriginalEdges + re.OriginalEdges,
				LeftChild:     fe.Id,
				RightChild:    re.Id,
			})
		}
	}

	for _, sc := range newShortcuts {
		// create and stores the new shortcuts. Update parent field of child shortcuts also

	}

	return Importance{
		edgeDiff:     edgeDiff,
		contractCost: contrCost,
	}, nil
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

func (d *Domain) shortestPath(src GeoPointId, dst GeoPointId, transport string) ([][]GeoPointId, error) {
	// bidirectional dijkstra
	fwdDist := ds.NewMap[GeoPointId, float64]()
	fwdParent := ds.NewMap[GeoPointId, GeoPointId]()
	fwdVisited := ds.NewSet[GeoPointId]()
	fwdPq := ds.NewPriorityQueue[vertexByDistance, GeoPointId]()
	fwdPq.Push(vertexByDistance{
		dist: 0,
		id:   src,
	})

	revDist := ds.NewMap[GeoPointId, float64]()
	revParent := ds.NewMap[GeoPointId, GeoPointId]()
	revVisited := ds.NewSet[GeoPointId]()
	revPq := ds.NewPriorityQueue[vertexByDistance, GeoPointId]()
	revPq.Push(vertexByDistance{
		dist: 0,
		id:   dst,
	})

	// bidirectional dijkstra
	var commonVertex []GeoPointId
	minDist := math.MaxFloat64
	for i := 0; !fwdPq.Empty() || !revPq.Empty(); i++ {
		pq1 := fwdPq
		d1 := fwdDist
		v1 := fwdVisited
		p1 := fwdParent
		d2 := revDist
		v2 := revVisited
		if i%2 == 1 {
			pq1 = revPq
			d1 = revDist
			v1 = revVisited
			p1 = revParent
			d2 = fwdDist
			v2 = fwdVisited
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
				commonVertex = nil
				minDist = v.dist + d2.Get(v.id)
			}
			if v.dist+d2.Get(v.id) == minDist {
				commonVertex = append(commonVertex, v.id)
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
			p1.Put(e.EndPoint2, v.id)
			pq1.Push(vertexByDistance{
				id:   e.EndPoint2,
				dist: newDist})
		}
	}
	if len(commonVertex) == 0 {
		return nil, nil
	}

	// path
	var paths [][]GeoPointId
	for _, cv := range commonVertex {
		var path []GeoPointId
		for id := cv; id != src; {
			parent := fwdParent.Get(id)
			path = append(path, parent)
			id = parent
		}
		reverse(path)

		for id := cv; id != src; {
			parent := revParent.Get(id)
			path = append(path, parent)
			id = parent
		}

		// resolve shortcut
		var unpackedPath []GeoPointId
		for i := 0; i < len(path)-1; i++ {
			unpack, err := d.GeoRepo.ResolveEdge(path[i], path[i+1], transport)
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
