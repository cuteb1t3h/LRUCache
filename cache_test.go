package lrucache

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestCacheInit(t *testing.T) {
	capacity := 10
	expected := 0

	cache := NewCache(capacity)

	assert.Equal(t, expected, cache.Len())
	assert.Equal(t, capacity, cache.Cap())
	assert.Nil(t, cache.head)
	assert.Nil(t, cache.tail)
}

func TestGet(t *testing.T) {
	capacity := 10
	cache := NewCache(capacity)

	cache.Add(1, "a")
	cache.Add(2, "b")

	val, ok := cache.Get(1)

	assert.Equal(t, "a", val)
	assert.True(t, ok)
}

func TestGetNotFound(t *testing.T) {
	capacity := 10
	cache := NewCache(capacity)

	cache.Add(2, "b")

	val, ok := cache.Get(1)

	assert.Equal(t, nil, val)
	assert.False(t, ok)
}

func TestAdd(t *testing.T) {
	capacity := 10
	cache := NewCache(capacity)

	cache.Add(2, "b")
	val, ok := cache.items[2]

	assert.True(t, ok)
	assert.Equal(t, "b", val.value)
	assert.Equal(t, 2, val.key)
	assert.Equal(t, val, cache.head)
	assert.Equal(t, val, cache.tail)
}

func TestAddRewrite(t *testing.T) {
	capacity := 10
	cache := NewCache(capacity)

	cache.Add(2, "b")
	node := cache.items[2]

	cache.Add(3, "c")

	cache.Add(2, "a")

	assert.Equal(t, cache.head, node)
	assert.Equal(t, node, cache.items[2])
	assert.Equal(t, 2, node.key)
}

func TestAddDelete(t *testing.T) {
	capacity := 2
	cache := NewCache(capacity)

	cache.Add(1, "a")
	cache.Add(2, "b")
	cache.Add(3, "c")

	head := cache.head.value
	tail := cache.tail.value

	_, ok := cache.items[1]
	assert.False(t, ok)

	_, ok = cache.items[3]
	assert.True(t, ok)

	assert.Equal(t, cache.Len(), cache.Cap())
	assert.Equal(t, "c", head)
	assert.Equal(t, "b", tail)
}

func TestAddWithTTL(t *testing.T) {
	capacity := 10
	cache := NewCache(capacity)

	cache.Add(2, "b")
	val, ok := cache.items[2]

	assert.True(t, ok)
	assert.Equal(t, "b", val.value)
	assert.Equal(t, 2, val.key)
	assert.Equal(t, val, cache.head)
	assert.Equal(t, val, cache.tail)
}

func TestAddWithTTLCheckDeleted(t *testing.T) {
	capacity := 3
	cache := NewCache(capacity)

	cache.AddWithTTL(1, "a", 5*time.Second)
	cache.AddWithTTL(2, "b", 10*time.Second)
	cache.AddWithTTL(3, "c", 3*time.Second)

	val, ok := cache.items[1]
	assert.True(t, ok)
	assert.Equal(t, 1, val.key)
	assert.Equal(t, "a", val.value)
	val, ok = cache.items[2]
	assert.True(t, ok)
	assert.Equal(t, 2, val.key)
	assert.Equal(t, "b", val.value)
	val, ok = cache.items[3]
	assert.True(t, ok)
	assert.Equal(t, 3, val.key)
	assert.Equal(t, "c", val.value)

	time.Sleep(5 * time.Second)

	w := sync.WaitGroup{}
	w.Add(3)
	go func() {
		defer w.Done()
		time.Sleep(5 * time.Second)
		_, ok := cache.Get(1)
		assert.False(t, ok)
	}()

	go func() {
		defer w.Done()
		time.Sleep(10 * time.Second)
		_, ok := cache.Get(2)
		assert.False(t, ok)
	}()

	go func() {
		defer w.Done()
		time.Sleep(3 * time.Second)
		_, ok := cache.Get(3)
		assert.False(t, ok)
	}()
	w.Wait()
}

func TestRemove(t *testing.T) {
	capacity := 3
	cache := NewCache(capacity)

	cache.Add(1, "a")
	cache.Add(2, "b")
	cache.Add(3, "c")

	length := cache.Len()
	cache.Remove(2)
	assert.Equal(t, length-1, cache.Len())

	_, ok := cache.Get(2)
	assert.False(t, ok)

	flag := true
	head := cache.head
	for head != cache.tail {
		head = head.next
		if head.key == 2 {
			flag = false
		}
	}
	assert.True(t, flag)

	_, ok = cache.items[2]
	assert.False(t, ok)
}

func TestClear(t *testing.T) {
	capacity := 3
	cache := NewCache(capacity)

	cache.Add(1, "a")
	cache.Add(2, "b")
	cache.Add(3, "c")

	cache.Clear()

	assert.Nil(t, cache.head)
	assert.Nil(t, cache.tail)
	assert.Equal(t, 0, cache.Len())
}
