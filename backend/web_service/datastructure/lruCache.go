package datastructure

type LruCache[K comparable, V any] struct {
	capacity     int
	size         int
	m            map[K]*listNode[K, V]
	sentinelHead *listNode[K, V]
	sentinelTail *listNode[K, V]
}

type listNode[K comparable, V any] struct {
	key  K
	val  V
	prev *listNode[K, V]
	next *listNode[K, V]
}

func (c *LruCache[K, V]) get(key K) (V, bool) {
	node, ok := c.m[key]
	var ret V
	if !ok {
		return ret, false
	}
	c.remove(node)
	c.makeHead(node)
	return node.val, true
}

func (c *LruCache[K, V]) put(key K, val V) {
	node, ok := c.m[key]
	if !ok {
		node = &listNode[K, V]{key: key, val: val}
		c.size++
		c.m[key] = node
	} else {
		c.remove(node)
	}
	c.makeHead(node)
	if c.size <= c.capacity {
		return
	}
	last := c.sentinelTail.prev
	c.remove(last)
	c.size--
	delete(c.m, last.key)
}

func (c *LruCache[K, V]) makeHead(node *listNode[K, V]) {
	node.next = c.sentinelHead.next
	node.prev = c.sentinelHead
	c.sentinelHead.next.prev = node
	c.sentinelHead.next = node
}

func (c *LruCache[K, V]) remove(node *listNode[K, V]) {
	node.next.prev = node.prev
	node.prev.next = node.next
}

func NewLruCache[K comparable, V any](cap int) *LruCache[K, V] {
	if cap == 0 {
		panic("LRU caches must have positive capacity")
	}
	var ret LruCache[K, V]
	ret.sentinelHead = &listNode[K, V]{}
	ret.sentinelTail = &listNode[K, V]{}
	ret.sentinelHead.next = ret.sentinelTail
	ret.sentinelTail.prev = ret.sentinelHead
	ret.capacity = cap
	return &ret
}
