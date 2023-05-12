package datastructures

type Queue[T any] struct {
	popStack  []T
	pushStack []T
	sz        int
}

func (q *Queue[T]) Push(val T) {
	q.pushStack = append(q.pushStack, val)
	q.sz++
}

func (q *Queue[T]) Pop() (T, bool) {
	ret, ok := q.Peek()
	if !ok {
		return ret, ok
	}
	q.popStack = q.popStack[:len(q.popStack)-1]
	q.sz--
	return ret, ok
}

func (q *Queue[T]) Peek() (T, bool) {
	var ret T
	if q.sz == 0 {
		return ret, false
	}
	if len(q.popStack) == 0 {
		for i := len(q.pushStack) - 1; i > -1; i-- {
			q.popStack = append(q.popStack, q.pushStack[i])
		}
		q.pushStack = []T{}
	}
	ret = q.popStack[len(q.popStack)-1]
	return ret, true
}

func (q *Queue[T]) IsEmpty() bool {
	return q.sz == 0
}

func (q *Queue[T]) Size() int {
	return q.sz
}
