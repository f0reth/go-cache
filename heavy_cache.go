package cache

import (
	"sync"
)

type HeavyCache[K comparable, V any] struct {
	mu    sync.RWMutex
	items map[K]V
}

// 新しいHeavyCacheインスタンスを作成します
// HeavyCacheはsync.RWMutexを使用し、読み取り操作の並行性を向上させます
func NewHeavy[K comparable, V any]() *HeavyCache[K, V] {
	return &HeavyCache[K, V]{
		items: make(map[K]V),
	}
}

// キャッシュに値を格納します
func (c *HeavyCache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	c.items[key] = value
	c.mu.Unlock()
}

// キャッシュから値を取得します
// 2つ目の戻り値は、キーが存在するかどうかを示します
func (c *HeavyCache[K, V]) Get(key K) (zero V, ok bool) {
	c.mu.RLock()
	item, ok := c.items[key]
	c.mu.RUnlock()
	return item, ok
}

// キャッシュから項目を削除します
// キーが存在しない場合は何もしません
func (c *HeavyCache[K, V]) Delete(key K) {
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
}

// キャッシュからすべての項目を削除します
func (c *HeavyCache[K, V]) Clear() {
	c.mu.Lock()
	c.items = make(map[K]V)
	c.mu.Unlock()
}

// キャッシュに格納されているアイテム数を返します
func (c *HeavyCache[K, V]) Len() int {
	c.mu.RLock()
	n := len(c.items)
	c.mu.RUnlock()
	return n
}

// キーがキャッシュに存在するかどうかを返します
func (c *HeavyCache[K, V]) Has(key K) bool {
	c.mu.RLock()
	_, ok := c.items[key]
	c.mu.RUnlock()
	return ok
}

// キャッシュに格納されているすべてのキーを返します
func (c *HeavyCache[K, V]) Keys() []K {
	c.mu.RLock()
	keys := make([]K, 0, len(c.items))
	for k := range c.items {
		keys = append(keys, k)
	}
	c.mu.RUnlock()
	return keys
}

// キャッシュに格納されているすべての値を返します
func (c *HeavyCache[K, V]) Values() []V {
	c.mu.RLock()
	values := make([]V, 0, len(c.items))
	for _, v := range c.items {
		values = append(values, v)
	}
	c.mu.RUnlock()
	return values
}

// キーが存在しない場合のみ値をセットします
// セットした場合はtrueを、既に存在した場合はfalseを返します
func (c *HeavyCache[K, V]) SetIfAbsent(key K, value V) bool {
	c.mu.Lock()
	_, exists := c.items[key]
	if !exists {
		c.items[key] = value
	}
	c.mu.Unlock()
	return !exists
}

// キーが存在すればその値を返し、存在しなければvalueをセットして返します
func (c *HeavyCache[K, V]) GetOrSet(key K, value V) V {
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
func (c *HeavyCache[K, V]) Snapshot() map[K]V {
	c.mu.RLock()
	result := make(map[K]V, len(c.items))
	for k, v := range c.items {
		result[k] = v
	}
	c.mu.RUnlock()
	return result
}

// キャッシュ内の全アイテムを巡回し、fnを呼び出します
// fnがfalseを返した場合、巡回を即座に停止します
// 注意: 読み取りロックを保持したまま全件走査するため、アイテム数が多い場合やfnが重い場合は書き込み操作をブロックします
func (c *HeavyCache[K, V]) Range(fn func(key K, value V) bool) {
	c.mu.RLock()
	for k, v := range c.items {
		if !fn(k, v) {
			break
		}
	}
	c.mu.RUnlock()
}

// 条件fnを満たすアイテムの数を返します
// 注意: 読み取りロックを保持したまま全件走査するため、アイテム数が多い場合やfnが重い場合は書き込み操作をブロックします
func (c *HeavyCache[K, V]) Count(fn func(K, V) bool) int {
	c.mu.RLock()
	n := 0
	for k, v := range c.items {
		if fn(k, v) {
			n++
		}
	}
	c.mu.RUnlock()
	return n
}

// キャッシュからすべての項目を取り出し、空にします
// 取り出した項目をマップとして返します
func (c *HeavyCache[K, V]) Drain() map[K]V {
	c.mu.Lock()
	items := c.items
	c.items = make(map[K]V)
	c.mu.Unlock()
	return items
}

// 条件fnを満たすアイテムを一括削除します
// 注意: ロックを保持したまま全件走査するため、アイテム数が多い場合やfnが重い場合は他の操作をブロックします
func (c *HeavyCache[K, V]) DeleteFunc(fn func(K, V) bool) {
	c.mu.Lock()
	for k, v := range c.items {
		if fn(k, v) {
			delete(c.items, k)
		}
	}
	c.mu.Unlock()
}

// キーの値を取得すると同時にキャッシュから削除します
func (c *HeavyCache[K, V]) Pop(key K) (zero V, ok bool) {
	c.mu.Lock()
	val, ok := c.items[key]
	if ok {
		delete(c.items, key)
	}
	c.mu.Unlock()
	return val, ok
}

// 複数アイテムを一度のロックでセットします
func (c *HeavyCache[K, V]) SetAll(items map[K]V) {
	c.mu.Lock()
	for k, v := range items {
		c.items[k] = v
	}
	c.mu.Unlock()
}

// 複数キーを一度のロックで取得します
// 存在するキーのみ結果に含まれます
func (c *HeavyCache[K, V]) GetAll(keys ...K) map[K]V {
	c.mu.RLock()
	result := make(map[K]V, len(keys))
	for _, k := range keys {
		if v, ok := c.items[k]; ok {
			result[k] = v
		}
	}
	c.mu.RUnlock()
	return result
}

// キーが存在すればその値を返し、存在しなければfn()の結果をセットして返します
// GetOrSetと違い、値の生成を遅延できるため、コストが高い初期化に適しています
// 注意: fnはロックを保持したまま実行されるため、fnが重い場合は他の操作をブロックします
func (c *HeavyCache[K, V]) GetOrSetFunc(key K, fn func() V) V {
	c.mu.Lock()
	if existing, ok := c.items[key]; ok {
		c.mu.Unlock()
		return existing
	}
	val := fn()
	c.items[key] = val
	c.mu.Unlock()
	return val
}

// キーが存在する場合、fnで値を更新します
// キーが存在した場合はtrueを、存在しなかった場合はfalseを返します
func (c *HeavyCache[K, V]) Update(key K, fn func(V) V) bool {
	c.mu.Lock()
	val, ok := c.items[key]
	if ok {
		c.items[key] = fn(val)
	}
	c.mu.Unlock()
	return ok
}

// 値を入れ替え、古い値を返します
// キーが存在しなかった場合は新しい値をセットし、ゼロ値とfalseを返します
func (c *HeavyCache[K, V]) Swap(key K, value V) (zero V, ok bool) {
	c.mu.Lock()
	old, ok := c.items[key]
	c.items[key] = value
	c.mu.Unlock()
	return old, ok
}

// oldとeqで比較して一致する場合のみnewに置き換えます
// キーが存在しない場合、または値が一致しない場合はfalseを返します
func (c *HeavyCache[K, V]) CompareAndSwap(key K, old, new V, eq func(V, V) bool) bool {
	c.mu.Lock()
	current, ok := c.items[key]
	if ok && eq(current, old) {
		c.items[key] = new
		c.mu.Unlock()
		return true
	}
	c.mu.Unlock()
	return false
}

// キーが既に存在する場合のみ値を上書きします
// 上書きした場合はtrueを、キーが存在しなかった場合はfalseを返します
func (c *HeavyCache[K, V]) Replace(key K, value V) bool {
	c.mu.Lock()
	_, ok := c.items[key]
	if ok {
		c.items[key] = value
	}
	c.mu.Unlock()
	return ok
}

// 全アイテムの値をfnの戻り値でインプレース変換します
// 注意: ロックを保持したまま全件走査するため、アイテム数が多い場合やfnが重い場合は他の操作をブロックします
func (c *HeavyCache[K, V]) Map(fn func(K, V) V) {
	c.mu.Lock()
	for k, v := range c.items {
		c.items[k] = fn(k, v)
	}
	c.mu.Unlock()
}

// 複数キーを一度のロックで削除します
func (c *HeavyCache[K, V]) DeleteAll(keys ...K) {
	c.mu.Lock()
	for _, k := range keys {
		delete(c.items, k)
	}
	c.mu.Unlock()
}

// 条件fnを満たすアイテムのスナップショットを返します
// キャッシュ自体は変更されません
// 注意: 読み取りロックを保持したまま全件走査するため、アイテム数が多い場合やfnが重い場合は書き込み操作をブロックします
func (c *HeavyCache[K, V]) Filter(fn func(K, V) bool) map[K]V {
	c.mu.RLock()
	result := make(map[K]V)
	for k, v := range c.items {
		if fn(k, v) {
			result[k] = v
		}
	}
	c.mu.RUnlock()
	return result
}

// 値がfnの条件を満たす場合のみ削除します
// 削除した場合はtrueを、キーが存在しないか条件を満たさない場合はfalseを返します
func (c *HeavyCache[K, V]) CompareAndDelete(key K, fn func(V) bool) bool {
	c.mu.Lock()
	val, ok := c.items[key]
	if ok && fn(val) {
		delete(c.items, key)
		c.mu.Unlock()
		return true
	}
	c.mu.Unlock()
	return false
}
