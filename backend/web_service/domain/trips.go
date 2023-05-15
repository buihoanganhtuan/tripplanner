package domain

import (
	"sort"
	"strings"

	"github.com/buihoanganhtuan/tripplanner/backend/web_service/datastructure"
)

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

type TripService interface {
	GetTrip(id TripId) (Trip, error)
	CreateTrip(t Trip) (Trip, error)
	ListTrips() ([]Trip, error)
	UpdateTrip(t Trip) (Trip, error)
	ReplaceTrip(t Trip) (Trip, error)
	PlanTrip(id TripId) (Trip, error)
	DeleteTrip(id TripId) error
}

type TripId string

// internal types, not exposed
type graphError []PointId
type cycleError []graphError
type multiFirstError graphError
type multiLastError graphError
type firstAndLastError graphError
type unknownIdError graphError
type pointOrder []PointId
type cycle []int

func (ge graphError) Error() string {
	var pids []string
	for _, pid := range ge {
		pids = append(pids, string(pid))
	}
	return strings.Join(pids, ",")
}

func (mf multiFirstError) Error() string {
	return graphError(mf).Error()
}

func (ml multiLastError) Error() string {
	return graphError(ml).Error()
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

func (sm firstAndLastError) Error() string {
	return graphError(sm).Error()
}

/*
Extract possible solutions to a certain DAG ordering. As the number of solutions can be
quite large, we terminate the search when the number of results found thus far exceed lim
*/

func topologicalSort(points []Point, startTime DateTime, lim int) ([]pointOrder, error) {
	intIds := datastructure.NewMap[PointId, int]()
	pointIds := datastructure.NewMap[int, PointId]()

	mapback := func(intIds []int) []PointId {
		var ids []PointId
		for _, id := range intIds {
			ids = append(ids, pointIds.Get(id))
		}
		return ids
	}

	// convert planner.PointId (string) to integer id
	var first, last, both []int
	for i, p := range points {
		intIds.Put(p.Id, i)
		pointIds.Put(i, p.Id)
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
		return nil, multiFirstError(mapback(first))
	}

	if len(last) > 1 {
		return nil, multiLastError(mapback(last))
	}

	if len(both) > 0 {
		return nil, multiLastError(mapback(both))
	}

	// verify that there's no unknown node
	var unknown []PointId
	for _, p := range points {
		if intIds.Exist(p.Id) {
			continue
		}
		unknown = append(unknown, p.Id)
	}
	if len(unknown) > 0 {
		return nil, unknownIdError(unknown)
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
	cycles := GetCycles(indeg, adj)
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

	var q, cur []int
	for i, d := range indeg {
		if d == 0 {
			q = append(q, i)
		}
	}
	dfs(q, cur, startTime)

	return res, nil
}

func GetCycles(indeg []int, adj [][]int) []cycle {
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
