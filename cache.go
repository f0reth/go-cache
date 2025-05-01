package cache

import (
	"sync"
)

type CacheInterface interface {
	Set(key any, value any)
	Get(key any) (any, bool)
	Delete(key any)
	Clear()
}

type Cache[K comparable, V any] struct {
	mu    sync.Mutex
	items map[K]V
}

// 新しいキャッシュインスタンスを作成します
func New[K comparable, V any]() *Cache[K, V] {
	return &Cache[K, V]{
		items: make(map[K]V),
	}
}

// キャッシュに値を格納します
func (c *Cache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = value
}

// キャッシュから値を取得します
// 2つ目の戻り値は、キーが存在するかどうかを示します
func (c *Cache[K, V]) Get(key K) (zero V, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, ok := c.items[key]
	if !ok {
		return zero, false
	}

	return item, ok
}

// キャッシュから項目を削除します
func (c *Cache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

// キャッシュからすべての項目を削除します
func (c *Cache[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[K]V)
}
