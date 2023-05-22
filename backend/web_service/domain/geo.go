package domain

import (
	"math"

	ds "github.com/buihoanganhtuan/tripplanner/backend/web_service/datastructure"
)

const (
	GeohashLen = 41
)

type GeoPoint struct {
	Id      GeoPointId     `json:"id"`
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
	Id        GeoEdgeId
	EndPoint1 GeoPointId
	EndPoint2 GeoPointId
	Cost      []EdgeCost
}

type EdgeCost struct {
	Transport string
	Value     float64
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
		fwdEdges, err := d.GeoRepo.EdgesToHigherImportance(curVertex.Id)
		if err != nil {
			return err
		}
		revEdges, err := d.GeoRepo.EdgesFromHigherImportance(curVertex.Id)
		if err != nil {
			return err
		}

		var targets []vertexByDistance
		var ma float64
		for _, e := range fwdEdges {
			for _, c := range e.Cost {
				if c.Transport != transport {
					continue
				}
				targets = append(targets, vertexByDistance{id: e.EndPoint2, dist: c.Value})
				ma = math.Max(ma, c.Value)
				break
			}
		}

		var sources []vertexByDistance
		for _, e := range revEdges {
			for _, c := range e.Cost {
				if c.Transport != transport {
					continue
				}
				sources = append(sources, vertexByDistance{id: e.EndPoint1, dist: c.Value})
				break
			}
		}

		for _, src := range sources {
			limit := ma + src.dist
			pq := ds.NewPriorityQueue[vertexByDistance, GeoPointId]()
			dist := ds.NewMap[GeoPointId, float64]()
			pq.Push(vertexByDistance{id: src.id, dist: 0})
			for !pq.Empty() {
				v := pq.Poll()
				// optimization 1: early termination
				if v.dist > limit {
					break
				}
				dist.Put(v.id, v.dist)
				edges, err := d.GeoRepo.EdgesToHigherImportance(v.id)
				if err != nil {
					return err
				}
				for _, e := range edges {
					if e.EndPoint2 == curVertex.Id {
						continue
					}
					for _, c := range e.Cost {
						if c.Transport != transport {
							continue
						}
						newDist := v.dist + c.Value
						if dist.Exist(e.EndPoint2) && dist.Get(e.EndPoint2) <= newDist {
							continue
						}
						dist.Put(e.EndPoint2, newDist)
						pq.Push(vertexByDistance{id: e.EndPoint2, dist: newDist})
					}
				}
			}
			// find the target of the shortcut from src
			var diff int
			for _, tgt := range targets {
				if dist.Exist(tgt.id) && src.dist+tgt.dist >= dist.Get(tgt.id) {
					continue
				}
				diff++
			}
			importPq.Push(vertexByImportance{id: curVertex.Id, importance: float64(diff - len(sources) - len(targets))})
		}
	}

	// start contraction
	for !importPq.Empty() {
		curVertex := importPq.Peek()
		// do local search

	}

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
			edges, err = d.GeoRepo.EdgesToHigherImportance(v.id)
		} else {
			edges, err = d.GeoRepo.EdgesFromHigherImportance(v.id)
		}

		if err != nil {
			return nil, err
		}
		for _, e := range edges {
			for _, c := range e.Cost {
				if c.Transport != transport {
					continue
				}
				if !v1.Contains(e.EndPoint2) && (!d1.Exist(e.EndPoint2) || d1.Get(e.EndPoint2) > v.dist+c.Value) {
					d1.Put(e.EndPoint2, v.dist+c.Value)
					p1.Put(e.EndPoint2, v.id)
					pq1.Push(vertexByDistance{id: e.EndPoint2, dist: v.dist + c.Value})
				}
			}
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
			unpack, err := d.GeoRepo.ResolveEdge(path[i], path[i+1])
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
