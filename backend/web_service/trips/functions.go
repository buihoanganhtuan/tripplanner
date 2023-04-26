package trips

import (
	"fmt"
	"net/http"

	cst "github.com/buihoanganhtuan/tripplanner/backend/web_service/_constants"
)

func newServerParseError() *cst.ErrorResponse {
	return &cst.ErrorResponse{
		Code:    http.StatusInternalServerError,
		Message: "server-side parsing error",
	}
}

func newClientParseError(field string) *cst.ErrorResponse {
	return &cst.ErrorResponse{
		Code:    http.StatusBadRequest,
		Message: fmt.Sprintf("fail to parse field %s", field),
	}
}

func newDatabaseQueryError() *cst.ErrorResponse {
	return &cst.ErrorResponse{
		Code:    http.StatusInternalServerError,
		Message: "database query error",
	}
}

func newUnknownError() *cst.ErrorResponse {
	return &cst.ErrorResponse{
		Code:    http.StatusInternalServerError,
		Message: "unknown server-side error",
	}
}

func topologicalSort(nodes []Node, lim int) (PlanResults, error) {
	nameMap := map[string]int{}
	nameRevMap := map[int]string{}

	var firstNode, lastNode, firstAndLast []int
	for i, node := range nodes {
		nameMap[node.Id] = i
		nameRevMap[i] = node.Id
		if node.First && node.Last {
			firstAndLast = append(firstAndLast, i)
		}
		if node.First {
			firstNode = append(firstNode, i)
		}
		if node.Last {
			lastNode = append(lastNode, i)
		}
	}

	backwardMap := func(intIds []int) []string {
		var ids []string
		for _, id := range intIds {
			ids = append(ids, nameRevMap[id])
		}
		return ids
	}

	if len(firstNode) > 1 {
		return nil, MultiFirstError(backwardMap(firstNode))
	}

	if len(lastNode) > 1 {
		return nil, MultiLastError(backwardMap(lastNode))
	}

	if len(firstAndLast) > 0 {
		return nil, MultiLastError(backwardMap(firstAndLast))
	}

	var unknownNodeIds []string

	// BFS
	indeg := make([]int, len(nameMap))
	adj := make([][]int, len(nameMap))
	for i, node := range nodes {
		for _, prev := range node.Before {
			j, ok := nameMap[prev]
			if !ok {
				unknownNodeIds = append(unknownNodeIds, prev)
				continue
			}
			indeg[i]++
			adj[j] = append(adj[j], i)
		}
		for _, next := range node.After {
			j, ok := nameMap[next]
			if !ok {
				unknownNodeIds = append(unknownNodeIds, next)
				continue
			}
			indeg[j]++
			adj[i] = append(adj[i], j)
		}
	}

	if len(unknownNodeIds) > 0 {
		return nil, UnknownNodeIdError(unknownNodeIds)
	}

	// Check for cycle
	cycles := GetCycles(indeg, adj)
	if cycles != nil {
		var ce []GraphError
		for _, c := range cycles {
			ce = append(ce, backwardMap(c))
		}
		return nil, CycleError(ce)
	}

	// O(N^2) because we want to check for more than one route

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
