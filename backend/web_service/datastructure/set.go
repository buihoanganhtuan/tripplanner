package datastructure

import "strings"

type Set[T comparable] struct {
	m  map[T]bool
	sz int
}

func (s *Set[T]) Add(val T) bool {
	_, ok := s.m[val]
	if ok {
		return false
	}
	s.m[val] = true
	s.sz++
	return true
}

func (s *Set[T]) AddAll(vals ...T) []bool {
	res := make([]bool, len(vals))
	for i, v := range vals {
		res[i] = s.Add(v)
	}
	return res
}

func (s *Set[T]) Remove(val T) bool {
	_, ok := s.m[val]
	if !ok {
		return false
	}
	delete(s.m, val)
	s.sz--
	return true
}

func (s *Set[T]) Contains(val T) bool {
	_, ok := s.m[val]
	return ok
}

func (s *Set[T]) Empty() bool {
	return s.sz == 0
}

func (s *Set[T]) Size() int {
	return s.sz
}

func (s *Set[T]) Values() []T {
	var res []T
	for k := range s.m {
		res = append(res, k)
	}
	return res
}

func (s *Set[T]) Intersection(os *Set[T]) *Set[T] {
	s1 := s
	s2 := os
	if s1.Size() > s2.Size() {
		s1, s2 = s2, s1
	}
	var res Set[T]
	for k := range s1.m {
		if !s2.Contains(k) {
			continue
		}
		res.Add(k)
	}
	return &res
}

func (s *Set[T]) Difference(os *Set[T]) *Set[T] {
	var res Set[T]
	for k := range s.m {
		if os.Contains(k) {
			continue
		}
		res.Add(k)
	}
	return &res
}

func (s *Set[T]) IsSubset(os *Set[T]) bool {
	for k := range s.m {
		if os.Contains(k) {
			continue
		}
		return false
	}
	return true
}

func (s *Set[T]) ToString(fmtFn func(T) string, sep string) string {
	tmp := make([]string, s.sz)
	var i int
	for k := range s.m {
		tmp[i] = fmtFn(k)
	}
	return strings.Join(tmp, sep)
}

func NewSet[T comparable]() *Set[T] {
	return &Set[T]{}
}

func NewDefaultSet[T comparable](vals ...T) *Set[T] {
	s := Set[T]{}
	for _, v := range vals {
		s.Add(v)
	}
	return &s
}
