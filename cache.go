package lrucache

import (
	"sync"
	"time"
)

// Node - Элемент двусвязного списка
type Node struct {
	key        any
	value      any
	exp        time.Time
	prev, next *Node
}

type LRUCache struct {
	items map[any]*Node
	cap   int
	head  *Node
	tail  *Node
	mutex sync.RWMutex
}

type ICache interface {
	Cap() int
	Len() int
	Clear()
	Add(key, value any)
	AddWithTTL(key, value any, ttl time.Duration)
	Get(key any) (value any, ok bool)
	Remove(key any)
}

// NewCache - Конструктор кеша
func NewCache(cap int) *LRUCache {
	cache := &LRUCache{
		cap:   cap,
		items: make(map[any]*Node, cap),
	}
	go cache.checkExpired()
	return cache
}

func (cache *LRUCache) Cap() int {
	c := cache.cap
	return c
}

func (cache *LRUCache) Len() int {
	cache.mutex.RLock()
	l := len(cache.items)
	cache.mutex.RUnlock()
	return l
}

// Clear - Отчистка всего кеша
func (cache *LRUCache) Clear() {
	cache.mutex.Lock()
	clear(cache.items)

	for cache.head != cache.tail {
		cache.head = cache.head.next
		cache.head.prev = nil
	}
	cache.head = nil
	cache.tail = nil
	cache.mutex.Unlock()
}

func (cache *LRUCache) Add(key, value any) {
	cache.add(key, value, 0)
}

// AddWithTTL - ttl должен быть больше 0
func (cache *LRUCache) AddWithTTL(key, value any, ttl time.Duration) {
	cache.add(key, value, ttl)
}

func (cache *LRUCache) add(key, value any, ttl time.Duration) {
	cache.mutex.Lock()

	if val, ok := cache.items[key]; !ok {
		node := &Node{key: key, value: value}
		if ttl != 0 {
			node.exp = time.Now().Add(ttl)
		}
		if len(cache.items) == 0 {
			cache.head = node
			cache.tail = node
		} else {
			// когда кеш полный, удаляется последний элемент для добавления нового
			if len(cache.items) == cache.cap {
				delete(cache.items, cache.tail.key)
				cache.tail.prev.next = nil
				cache.tail = cache.tail.prev
			}

			node.next = cache.head
			cache.head.prev = node
			cache.head = node
		}
		cache.items[key] = node
	} else {
		if ttl != 0 {
			val.exp = time.Now().Add(ttl)
		}
		val.value = value

		if cache.head != val {
			if val.prev != nil {
				val.prev.next = val.next
			}
			if val.next != nil {
				val.next.prev = val.prev
			}

			val.next = cache.head
			val.prev = nil
			cache.head.prev = val
			cache.head = val
		}
	}
	cache.mutex.Unlock()
}

// Get - Получение элемента по ключу
func (cache *LRUCache) Get(key any) (value any, ok bool) {
	cache.mutex.RLock()
	if val, ok := cache.items[key]; ok {
		cache.mutex.RUnlock()
		return val.value, ok
	} else {
		cache.mutex.RUnlock()
		return nil, ok
	}
}

// Remove - Удаление элемента по ключу
func (cache *LRUCache) Remove(key any) {
	cache.mutex.Lock()
	if val, ok := cache.items[key]; ok {
		if val.prev != nil {
			val.prev.next = val.next
		}
		if val.next != nil {
			val.next.prev = val.prev
		}

		delete(cache.items, key)
	}
	cache.mutex.Unlock()
}

// checkExpired - Проверяет истекло ли время хранения элемента кеша
func (cache *LRUCache) checkExpired() {
	for {
		time.Sleep(time.Second * 5)
		cache.mutex.Lock()
		for key, val := range cache.items {
			if !val.exp.IsZero() {
				if time.Now().After(val.exp) {
					if val.prev != nil {
						val.prev.next = val.next
					}
					if val.next != nil {
						val.next.prev = val.prev
					}

					delete(cache.items, key)
				}
			}
		}
		cache.mutex.Unlock()
	}
}
