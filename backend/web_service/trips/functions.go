package trips

import (
	"time"

	cst "github.com/buihoanganhtuan/tripplanner/backend/web_service/_constants"
	"github.com/golang-jwt/jwt/v4"
)

func createJsonTime(t *time.Time) Datetime {
	return Datetime{
		Year:  t.Year(),
		Month: int(t.Month()),
		Day:   t.Day(),
		Hour:  t.Hour(),
		Min:   t.Minute(),
	}
}

func verifyJwtClaims(cm jwt.MapClaims) (bool, string) {
	if !cm.VerifyIssuer(cst.AuthServiceName, true) {
		return false, "iss"
	}
	if !cm.VerifyAudience(cst.WebServiceName, true) {
		return false, "aud"
	}
	if !cm.VerifyExpiresAt(time.Now().Unix(), true) {
		return false, "exp"
	}
	// TODO: implement a custom JSON-valued claim to implement authorization
	// We have a set of (resource, method) pairs. Each security role refers to
	// a specific set of those pairs and indicates that the role can perform
	// those specific methods on those specific resources. An identity can assume
	// one or multiple roles, depending on our policy
}

func topologicalSort(nodes []Node, lim int) ([][]string, error) {
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
	outdeg := make([]int, len(nameMap))
	adj := make([][]int, len(nameMap))
	for i, node := range nodes {
		for _, prev := range node.Before {
			j, ok := nameMap[prev]
			if !ok {
				unknownNodeIds = append(unknownNodeIds, prev)
				continue
			}
			indeg[i]++
			outdeg[j]++
			adj[j] = append(adj[j], i)
		}
		for _, next := range node.After {
			j, ok := nameMap[next]
			if !ok {
				unknownNodeIds = append(unknownNodeIds, next)
				continue
			}
			indeg[j]++
			outdeg[i]++
			adj[i] = append(adj[i], j)
		}

	}

	if len(unknownNodeIds) > 0 {
		return nil, UnknownNodeIdError(unknownNodeIds)
	}

	// Check for cycle

	// O(N^2) because we want to check for more than one route
}

func CheckCycles(indeg []int, adj [][]int) [][]int {
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
	for i, d := range indegCp {
		if d > 0 {
			ok = false
		}
	}

	if ok {
		return nil
	}

	st = make([]int, 0)
	return GetAllCycles(adj)
}

func GetAllCycles(adj [][]int) [][]int {
	var res [][]int

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
				var cycle []int
				var add bool
				for _, node := range path {
					if node == j {
						add = true
					}
					if add {
						cycle = append(cycle, node)
					}
				}
				res = append(res, cycle)
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
