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
	SetIfAbsent(key K, value V) bool
	GetOrSet(key K, value V) V
	Snapshot() map[K]V
	Range(fn func(key K, value V) bool)
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

// キーが存在しない場合のみ値をセットします
// セットした場合はtrueを、既に存在した場合はfalseを返します
func (c *Cache[K, V]) SetIfAbsent(key K, value V) bool {
	c.mu.Lock()
	_, exists := c.items[key]
	if !exists {
		c.items[key] = value
	}
	c.mu.Unlock()
	return !exists
}

// キーが存在すればその値を返し、存在しなければvalueをセットして返します
func (c *Cache[K, V]) GetOrSet(key K, value V) V {
	c.mu.Lock()
	if existing, ok := c.items[key]; ok {
		c.mu.Unlock()
		return existing
	}
	c.items[key] = value
	c.mu.Unlock()
	return value
}

// キャッシュに格納されているすべてのアイテムのコピーを返します
// Drainと違い、キャッシュは変更されません
func (c *Cache[K, V]) Snapshot() map[K]V {
	c.mu.Lock()
	result := make(map[K]V, len(c.items))
	for k, v := range c.items {
		result[k] = v
	}
	c.mu.Unlock()
	return result
}

// キャッシュ内の全アイテムを巡回し、fnを呼び出します
// fnがfalseを返した場合、巡回を即座に停止します
func (c *Cache[K, V]) Range(fn func(key K, value V) bool) {
	c.mu.Lock()
	for k, v := range c.items {
		if !fn(k, v) {
			break
		}
	}
	c.mu.Unlock()
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
