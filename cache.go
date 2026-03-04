package cache

import (
	"sync"
)

type CacheInterface[K comparable, V any] interface {
	Set(key K, value V)
	Get(key K) (V, bool)
	Delete(key K)
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
	c.items[key] = value
	c.mu.Unlock()
}

// キャッシュから値を取得します
// 2つ目の戻り値は、キーが存在するかどうかを示します
func (c *Cache[K, V]) Get(key K) (zero V, ok bool) {
	c.mu.Lock()
	item, ok := c.items[key]
	c.mu.Unlock()
	return item, ok
}

// キャッシュから項目を削除します
func (c *Cache[K, V]) Delete(key K) {
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
}

// キャッシュからすべての項目を削除します
func (c *Cache[K, V]) Clear() {
	c.mu.Lock()
	c.items = make(map[K]V)
	c.mu.Unlock()
}
