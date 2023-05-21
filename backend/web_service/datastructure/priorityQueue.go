package datastructure

type IdentifiableOrdered[T any, I comparable] interface {
	Id() I
	Less(x T) bool
}

type PriorityQueue[T IdentifiableOrdered[T, I], I comparable] struct {
	arr []T
	idx map[I]int
}

func (pq *PriorityQueue[T, I]) Push(e T) {
	i, ok := pq.idx[e.Id()]
	if !ok {
		pq.arr = append(pq.arr, e)
		i = len(pq.arr) - 1
		pq.idx[e.Id()] = i
	}
	pq.siftUp(i)
	pq.siftDown(i)
}

func (pq *PriorityQueue[T, I]) Peek() T {
	if len(pq.arr) == 0 {
		panic("trying to peek an empty queue")
	}
	return pq.arr[0]
}

func (pq *PriorityQueue[T, I]) Poll() T {
	if len(pq.arr) == 0 {
		panic("trying to poll an empty queue")
	}
	ret := pq.arr[0]
	l := len(pq.arr)
	if l > 1 {
		pq.arr[0] = pq.arr[l-1]
		pq.arr = pq.arr[:l-1]
		pq.siftDown(0)
	} else {
		pq.arr = nil
	}

	return ret
}

func (pq *PriorityQueue[T, I]) Remove(id I) bool {
	_, ok := pq.idx[id]
	if ok {
		delete(pq.idx, id)
	}
	return ok
}

func (pq *PriorityQueue[T, I]) Exist(id I) bool {
	_, ok := pq.idx[id]
	return ok
}

func (pq *PriorityQueue[T, I]) Size() int {
	return len(pq.arr)
}

func (pq *PriorityQueue[T, I]) Empty() bool {
	return len(pq.arr) > 0
}

func (pq *PriorityQueue[T, I]) siftDown(i int) {
	// used during removal
	left := 2*i + 1
	right := 2*i + 2
	smallerChild := left
	if right < len(pq.arr) && pq.arr[right].Less(pq.arr[left]) {
		smallerChild = right
	}
	if smallerChild < len(pq.arr) && pq.arr[smallerChild].Less(pq.arr[i]) {
		pq.idx[pq.arr[i].Id()] = smallerChild
		pq.idx[pq.arr[smallerChild].Id()] = i
		pq.arr[i], pq.arr[smallerChild] = pq.arr[smallerChild], pq.arr[i]
		pq.siftDown(smallerChild)
	}
}

func (pq *PriorityQueue[T, I]) siftUp(i int) {
	// used during insertion
	parent := (i - 1) / 2
	if i%2 == 0 {
		parent = (i - 2) / 2
	}
	if parent > 0 && pq.arr[i].Less(pq.arr[parent]) {
		pq.idx[pq.arr[i].Id()] = parent
		pq.idx[pq.arr[parent].Id()] = i
		pq.arr[parent], pq.arr[i] = pq.arr[i], pq.arr[parent]
		pq.siftUp(parent)
	}
}

func NewPriorityQueue[T IdentifiableOrdered[T, I], I comparable]() *PriorityQueue[T, I] {
	return &PriorityQueue[T, I]{}
}
