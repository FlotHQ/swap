package lru

import (
	"container/list"
	"sync"
)

type Cache[K comparable, V any] struct {
	maxBytes    int
	currentSize int
	items       map[K]*list.Element
	lru         *list.List
	mutex       sync.RWMutex
	funcSize    func(K, V) int
}

type entry[K comparable, V any] struct {
	key   K
	value V
}

func New[K comparable, V any](maxBytes int, funcSize func(K, V) int) *Cache[K, V] {
	return &Cache[K, V]{
		maxBytes: maxBytes,
		items:    make(map[K]*list.Element),
		lru:      list.New(),
		funcSize: funcSize,
	}
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if element, found := c.items[key]; found {
		c.lru.MoveToFront(element)
		return element.Value.(*entry[K, V]).value, true
	}
	var zero V
	return zero, false
}

func (c *Cache[K, V]) Set(key K, value V) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	newSize := c.funcSize(key, value)

	if newSize > c.maxBytes {
		return
	}

	if element, found := c.items[key]; found {
		entry := element.Value.(*entry[K, V])
		c.currentSize += newSize - c.funcSize(entry.key, entry.value)
		entry.value = value
		c.lru.MoveToFront(element)
	} else {
		c.evict(newSize)
		element := c.lru.PushFront(&entry[K, V]{key, value})
		c.items[key] = element
		c.currentSize += newSize
	}
}

func (c *Cache[K, V]) evict(requiredSpace int) {
	for c.currentSize+requiredSpace > c.maxBytes && c.lru.Len() > 0 {
		c.removeOldest()
	}
}

func (c *Cache[K, V]) removeOldest() {
	element := c.lru.Back()
	if element != nil {
		c.removeElement(element)
	}
}

func (c *Cache[K, V]) removeElement(element *list.Element) {
	entry := element.Value.(*entry[K, V])
	c.lru.Remove(element)
	delete(c.items, entry.key)
	c.currentSize -= c.funcSize(entry.key, entry.value)
}

func (c *Cache[K, V]) Remove(key K) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if element, found := c.items[key]; found {
		c.removeElement(element)
	}
}

func (c *Cache[K, V]) Len() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.items)
}

func (c *Cache[K, V]) ByteSize() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.currentSize
}
