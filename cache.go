package cache

import (
	"sync"
)

type CacheInterface[K comparable, V any] interface {
	Set(key K, value V)
	Get(key K) (V, bool)
	Delete(key K)
	Clear()
	Drain() map[K]V
	Len() int
	Has(key K) bool
	Keys() []K
	Values() []V
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
// キーが存在しない場合は何もしません
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

// キャッシュに格納されているアイテム数を返します
func (c *Cache[K, V]) Len() int {
	c.mu.Lock()
	n := len(c.items)
	c.mu.Unlock()
	return n
}

// キーがキャッシュに存在するかどうかを返します
func (c *Cache[K, V]) Has(key K) bool {
	c.mu.Lock()
	_, ok := c.items[key]
	c.mu.Unlock()
	return ok
}

// キャッシュに格納されているすべてのキーを返します
func (c *Cache[K, V]) Keys() []K {
	c.mu.Lock()
	keys := make([]K, 0, len(c.items))
	for k := range c.items {
		keys = append(keys, k)
	}
	c.mu.Unlock()
	return keys
}

// キャッシュに格納されているすべての値を返します
func (c *Cache[K, V]) Values() []V {
	c.mu.Lock()
	values := make([]V, 0, len(c.items))
	for _, v := range c.items {
		values = append(values, v)
	}
	c.mu.Unlock()
	return values
}

// キャッシュからすべての項目を取り出し、空にします
// 取り出した項目をマップとして返します
func (c *Cache[K, V]) Drain() map[K]V {
	c.mu.Lock()
	items := c.items
	c.items = make(map[K]V)
	c.mu.Unlock()
	return items
}
