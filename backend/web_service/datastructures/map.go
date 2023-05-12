package datastructures

type Map[K comparable, V any] struct {
	mp map[K]V
}

func (m *Map[K, V]) Put(key K, val V) V {
	v := m.mp[key]
	m.mp[key] = val
	return v
}

func (m *Map[K, V]) PutIfAbsent(key K, val V) V {
	v, ok := m.mp[key]
	if !ok {
		m.mp[key] = val
		return v
	}
	return val
}

func (m *Map[K, V]) Get(key K) V {
	return m.mp[key]
}

func (m *Map[K, V]) GetIfPresent(key K) (V, bool) {
	v, ok := m.mp[key]
	return v, ok
}

func (m *Map[K, V]) GetOrDefault(key K, def V) V {
	v, ok := m.mp[key]
	if ok {
		return v
	}
	return def
}

func (m *Map[K, V]) Remove(key K) bool {
	_, ok := m.mp[key]
	if ok {
		delete(m.mp, key)
		return true
	}
	return false
}

func (m *Map[K, V]) Keys() []K {
	var ks []K
	for k := range m.mp {
		ks = append(ks, k)
	}
	return ks
}

func (m *Map[K, V]) Values() []V {
	var vs []V
	for _, v := range m.mp {
		vs = append(vs, v)
	}
	return vs
}

func (m *Map[K, V]) Exist(key K) bool {
	_, ok := m.mp[key]
	return ok
}

func (m *Map[K, V]) Size() int {
	return len(m.mp)
}

func NewMap[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{
		mp: map[K]V{},
	}
}
