package domain

import (
	"github.com/buihoanganhtuan/tripplanner/backend/web_service/datastructure"
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
	Transport []string
	Cost      []float64
}

func (d *Domain) importanceUpdate() error {
	vertices, err := d.GeoRepo.ListTransitNodes()
	if err != nil {
		return err
	}
	for _, v := range vertices {
		edges, err := d.GeoRepo.EdgesToHigherOrEqualImportance(v)
		if err != nil {
			return err
		}

		var importance int
		for i := 0; i < len(edges); i++ {
			for j := i + 1; j < len(edges); j++ {
				// do normal dijkstra

			}
		}

	}
}

// Internal type used for shortest path finding
type vertex struct {
	dist float64
	id   GeoPointId
}

func (v vertex) Less(v2 vertex) bool {
	return v.dist < v2.dist
}

func (v vertex) Id() GeoPointId {
	return v.id
}

func (d *Domain) shortestPath(src GeoPointId, dst GeoPointId, transport string) ([]GeoPointId, error) {
	// bidirectional dijkstra
	fwdDist := datastructure.NewMap[GeoPointId, float64]()
	fwdParent := datastructure.NewMap[GeoPointId, GeoPointId]()
	fwdVisited := datastructure.NewSet[GeoPointId]()
	fwdPq := datastructure.NewPriorityQueue[vertex, GeoPointId]()
	fwdPq.Push(vertex{
		dist: 0,
		id:   src,
	})

	revDist := datastructure.NewMap[GeoPointId, float64]()
	revParent := datastructure.NewMap[GeoPointId, GeoPointId]()
	revVisited := datastructure.NewSet[GeoPointId]()
	revPq := datastructure.NewPriorityQueue[vertex, GeoPointId]()
	revPq.Push(vertex{
		dist: 0,
		id:   dst,
	})

	// stop if this is forward turn, and the current vertex is already in the visited set of the reverse routine and vice versa
	var commonVertex *GeoPointId
	for i := 0; !fwdPq.Empty() || !revPq.Empty(); i++ {
		pq1 := fwdPq
		d1 := fwdDist
		v1 := fwdVisited
		p1 := fwdParent
		v2 := revVisited
		if i%2 == 1 {
			pq1 = revPq
			d1 = revDist
			v1 = revVisited
			p1 = revParent
			v2 = fwdVisited
		}

		if pq1.Empty() {
			continue
		}
		v := pq1.Poll()
		if v1.Contains(v.id) {
			continue
		}
		if v2.Contains(v.id) {
			commonVertex = &v.id
			break
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
			for i := range e.Transport {
				if e.Transport[i] != transport {
					continue
				}
				if !v1.Contains(e.EndPoint2) && (!d1.Exist(e.EndPoint2) || d1.Get(e.EndPoint2) > v.dist+e.Cost[i]) {
					d1.Put(e.EndPoint2, v.dist+e.Cost[i])
					p1.Put(e.EndPoint2, v.id)
					pq1.Push(vertex{id: e.EndPoint2, dist: v.dist + e.Cost[i]})
				}
			}
		}
	}
	if commonVertex == nil {
		return nil, nil
	}

	// path
	var path []GeoPointId
	for id := *commonVertex; id != src; {
		parent := fwdParent.Get(id)
		path = append(path, parent)
		id = parent
	}
	reverse(path)

	for id := *commonVertex; id != src; {
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
