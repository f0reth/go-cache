package cache

import (
	"fmt"
	"sync"
	"testing"
)

func TestHeavyCacheNew(t *testing.T) {
	t.Parallel()

	c := NewHeavy[string, int]()
	if c == nil {
		t.Fatal("NewHeavy should return a non-nil cache")
	}
	if c.items == nil {
		t.Fatal("NewHeavy should initialize the items map")
	}
	if len(c.items) != 0 {
		t.Errorf("New cache should be empty, got size %d", len(c.items))
	}
}

func TestHeavyCacheSetAndGet(t *testing.T) {
	t.Parallel()

	c := NewHeavy[string, int]()

	// キーが存在しない場合
	val, ok := c.Get("not-exists")
	if ok {
		t.Error("Get should return false for non-existent keys")
	}
	if val != 0 {
		t.Errorf("Get should return zero value for non-existent keys, got %v", val)
	}

	// 値を設定
	c.Set("key1", 42)

	// 値を取得
	val, ok = c.Get("key1")
	if !ok {
		t.Error("Get should return true for existing keys")
	}
	if val != 42 {
		t.Errorf("Get should return the correct value, expected 42, got %v", val)
	}

	// 値を上書き
	c.Set("key1", 100)
	val, ok = c.Get("key1")
	if !ok {
		t.Error("Get should return true after overwriting a key")
	}
	if val != 100 {
		t.Errorf("Get should return the updated value, expected 100, got %v", val)
	}
}

func TestHeavyCacheDelete(t *testing.T) {
	t.Parallel()

	c := NewHeavy[string, int]()

	c.Set("key1", 42)
	c.Set("key2", 100)

	c.Delete("key1")

	val, ok := c.Get("key1")
	if ok {
		t.Error("Get should return false for deleted keys")
	}
	if val != 0 {
		t.Errorf("Get should return zero value for deleted keys, got %v", val)
	}

	val, ok = c.Get("key2")
	if !ok {
		t.Error("Delete should not affect other keys")
	}
	if val != 100 {
		t.Errorf("Get should return the correct value for non-deleted keys, expected 100, got %v", val)
	}

	// 存在しないキーを削除してもエラーにならない
	c.Delete("not-exists")
}

func TestHeavyCacheClear(t *testing.T) {
	t.Parallel()

	c := NewHeavy[string, int]()

	c.Set("key1", 42)
	c.Set("key2", 100)
	c.Set("key3", 200)

	c.Clear()

	for _, key := range []string{"key1", "key2", "key3"} {
		val, ok := c.Get(key)
		if ok {
			t.Errorf("Get should return false for key %s after Clear", key)
		}
		if val != 0 {
			t.Errorf("Get should return zero value for key %s after Clear, got %v", key, val)
		}
	}

	c.Set("new-key", 300)
	val, ok := c.Get("new-key")
	if !ok {
		t.Error("Set should work after Clear")
	}
	if val != 300 {
		t.Errorf("Get should return the correct value after Clear, expected 300, got %v", val)
	}
}

func TestHeavyCacheLen(t *testing.T) {
	t.Parallel()

	c := NewHeavy[string, int]()

	if n := c.Len(); n != 0 {
		t.Errorf("Len should return 0 for empty cache, got %d", n)
	}

	c.Set("key1", 1)
	if n := c.Len(); n != 1 {
		t.Errorf("Len should return 1 after one Set, got %d", n)
	}

	c.Set("key2", 2)
	c.Set("key3", 3)
	if n := c.Len(); n != 3 {
		t.Errorf("Len should return 3 after three Sets, got %d", n)
	}

	c.Set("key1", 100)
	if n := c.Len(); n != 3 {
		t.Errorf("Len should return 3 after overwriting a key, got %d", n)
	}

	c.Delete("key1")
	if n := c.Len(); n != 2 {
		t.Errorf("Len should return 2 after Delete, got %d", n)
	}

	c.Clear()
	if n := c.Len(); n != 0 {
		t.Errorf("Len should return 0 after Clear, got %d", n)
	}

	c.Set("key1", 1)
	c.Set("key2", 2)
	c.Drain()
	if n := c.Len(); n != 0 {
		t.Errorf("Len should return 0 after Drain, got %d", n)
	}
}

func TestHeavyCacheHas(t *testing.T) {
	t.Parallel()

	c := NewHeavy[string, int]()

	if c.Has("not-exists") {
		t.Error("Has should return false for non-existent keys")
	}

	c.Set("key1", 42)
	if !c.Has("key1") {
		t.Error("Has should return true for existing keys")
	}

	c.Delete("key1")
	if c.Has("key1") {
		t.Error("Has should return false after Delete")
	}

	cp := NewHeavy[string, *int]()
	cp.Set("nil-key", nil)
	if !cp.Has("nil-key") {
		t.Error("Has should return true for keys with nil values")
	}

	c.Set("key2", 100)
	c.Clear()
	if c.Has("key2") {
		t.Error("Has should return false after Clear")
	}
}

func TestHeavyCacheKeys(t *testing.T) {
	t.Parallel()

	c := NewHeavy[string, int]()

	keys := c.Keys()
	if len(keys) != 0 {
		t.Errorf("Keys should return empty slice for empty cache, got %v", keys)
	}

	c.Set("key1", 1)
	c.Set("key2", 2)
	c.Set("key3", 3)

	keys = c.Keys()
	if len(keys) != 3 {
		t.Errorf("Keys should return 3 keys, got %d", len(keys))
	}

	keySet := make(map[string]bool, len(keys))
	for _, k := range keys {
		keySet[k] = true
	}
	for _, expected := range []string{"key1", "key2", "key3"} {
		if !keySet[expected] {
			t.Errorf("Keys should contain %q", expected)
		}
	}

	c.Delete("key1")
	keys = c.Keys()
	if len(keys) != 2 {
		t.Errorf("Keys should return 2 keys after Delete, got %d", len(keys))
	}

	c.Clear()
	keys = c.Keys()
	if len(keys) != 0 {
		t.Errorf("Keys should return empty slice after Clear, got %v", keys)
	}
}

func TestHeavyCacheValues(t *testing.T) {
	t.Parallel()

	c := NewHeavy[string, int]()

	values := c.Values()
	if len(values) != 0 {
		t.Errorf("Values should return empty slice for empty cache, got %v", values)
	}

	c.Set("key1", 1)
	c.Set("key2", 2)
	c.Set("key3", 3)

	values = c.Values()
	if len(values) != 3 {
		t.Errorf("Values should return 3 values, got %d", len(values))
	}

	valSet := make(map[int]bool, len(values))
	for _, v := range values {
		valSet[v] = true
	}
	for _, expected := range []int{1, 2, 3} {
		if !valSet[expected] {
			t.Errorf("Values should contain %d", expected)
		}
	}

	c.Delete("key1")
	values = c.Values()
	if len(values) != 2 {
		t.Errorf("Values should return 2 values after Delete, got %d", len(values))
	}

	c.Clear()
	values = c.Values()
	if len(values) != 0 {
		t.Errorf("Values should return empty slice after Clear, got %v", values)
	}
}

func TestHeavyCacheSetIfAbsent(t *testing.T) {
	t.Parallel()

	c := NewHeavy[string, int]()

	if ok := c.SetIfAbsent("key1", 42); !ok {
		t.Error("SetIfAbsent should return true when key does not exist")
	}
	val, _ := c.Get("key1")
	if val != 42 {
		t.Errorf("SetIfAbsent should set the value, expected 42, got %v", val)
	}

	if ok := c.SetIfAbsent("key1", 999); ok {
		t.Error("SetIfAbsent should return false when key already exists")
	}
	val, _ = c.Get("key1")
	if val != 42 {
		t.Errorf("SetIfAbsent should not overwrite existing value, expected 42, got %v", val)
	}

	c.Delete("key1")
	if ok := c.SetIfAbsent("key1", 100); !ok {
		t.Error("SetIfAbsent should return true after Delete")
	}
	val, _ = c.Get("key1")
	if val != 100 {
		t.Errorf("SetIfAbsent should set new value after Delete, expected 100, got %v", val)
	}
}

func TestHeavyCacheGetOrSet(t *testing.T) {
	t.Parallel()

	c := NewHeavy[string, int]()

	val := c.GetOrSet("key1", 42)
	if val != 42 {
		t.Errorf("GetOrSet should return the set value, expected 42, got %v", val)
	}
	stored, ok := c.Get("key1")
	if !ok || stored != 42 {
		t.Errorf("GetOrSet should persist the value in cache, expected 42, got %v", stored)
	}

	val = c.GetOrSet("key1", 999)
	if val != 42 {
		t.Errorf("GetOrSet should return existing value, expected 42, got %v", val)
	}
	stored, _ = c.Get("key1")
	if stored != 42 {
		t.Errorf("GetOrSet should not overwrite existing value, expected 42, got %v", stored)
	}
}

func TestHeavyCacheSnapshot(t *testing.T) {
	t.Parallel()

	c := NewHeavy[string, int]()

	items := c.Snapshot()
	if len(items) != 0 {
		t.Errorf("Snapshot should return empty map for empty cache, got %v", items)
	}

	c.Set("key1", 1)
	c.Set("key2", 2)
	c.Set("key3", 3)

	items = c.Snapshot()
	if len(items) != 3 {
		t.Errorf("Snapshot should return 3 items, got %d", len(items))
	}
	if items["key1"] != 1 || items["key2"] != 2 || items["key3"] != 3 {
		t.Errorf("Snapshot should return correct values, got %v", items)
	}

	if c.Len() != 3 {
		t.Errorf("Snapshot should not remove items from cache, got len %d", c.Len())
	}

	items["key1"] = 999
	items["key4"] = 4
	val, ok := c.Get("key1")
	if !ok || val != 1 {
		t.Errorf("Modifying Snapshot map should not affect cache, expected 1, got %v", val)
	}
	if c.Len() != 3 {
		t.Errorf("Modifying Snapshot map should not change cache size, got %d", c.Len())
	}
}

func TestHeavyCacheRange(t *testing.T) {
	t.Parallel()

	c := NewHeavy[string, int]()

	// 空のキャッシュでRangeしても何も起きない
	count := 0
	c.Range(func(key string, value int) bool {
		count++
		return true
	})
	if count != 0 {
		t.Errorf("Range on empty cache should not call fn, got %d calls", count)
	}

	// 全アイテムを巡回
	c.Set("key1", 1)
	c.Set("key2", 2)
	c.Set("key3", 3)

	visited := make(map[string]int)
	c.Range(func(key string, value int) bool {
		visited[key] = value
		return true
	})
	if len(visited) != 3 {
		t.Errorf("Range should visit all 3 items, got %d", len(visited))
	}
	if visited["key1"] != 1 || visited["key2"] != 2 || visited["key3"] != 3 {
		t.Errorf("Range should visit correct key-value pairs, got %v", visited)
	}

	// falseを返すと即停止
	stopCount := 0
	c.Range(func(key string, value int) bool {
		stopCount++
		return false
	})
	if stopCount != 1 {
		t.Errorf("Range should stop after fn returns false, got %d calls", stopCount)
	}
}

func TestHeavyCacheRangeConcurrency(t *testing.T) {
	t.Parallel()

	c := NewHeavy[int, int]()
	var wg sync.WaitGroup

	for i := range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Set(i, i*10)
		}()
	}
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Range(func(key int, value int) bool {
				return true
			})
		}()
	}

	wg.Wait()
}

func TestHeavyCacheDrain(t *testing.T) {
	t.Parallel()

	c := NewHeavy[string, int]()
	c.Set("key1", 1)
	c.Set("key2", 2)
	c.Set("key3", 3)

	items := c.Drain()

	if len(items) != 3 {
		t.Errorf("Drain should return all items, expected 3, got %d", len(items))
	}
	if items["key1"] != 1 || items["key2"] != 2 || items["key3"] != 3 {
		t.Errorf("Drain should return correct values, got %v", items)
	}

	for _, key := range []string{"key1", "key2", "key3"} {
		_, ok := c.Get(key)
		if ok {
			t.Errorf("Cache should be empty after Drain, but key %s still exists", key)
		}
	}

	c.Set("new-key", 100)
	val, ok := c.Get("new-key")
	if !ok || val != 100 {
		t.Errorf("Cache should be reusable after Drain, got %v, %v", val, ok)
	}
}

func TestHeavyCacheDrainEmpty(t *testing.T) {
	t.Parallel()

	c := NewHeavy[string, int]()
	items := c.Drain()
	if len(items) != 0 {
		t.Errorf("Drain on empty cache should return empty map, got %d items", len(items))
	}
}

func TestHeavyCacheDrainIsolation(t *testing.T) {
	t.Parallel()

	c := NewHeavy[string, int]()
	c.Set("key1", 1)

	items := c.Drain()
	items["key1"] = 999
	items["key2"] = 2

	_, ok := c.Get("key1")
	if ok {
		t.Error("Modifying drained map should not affect cache")
	}
	if len(c.Drain()) != 0 {
		t.Error("Cache should remain empty after modifying drained map")
	}
}

func TestHeavyCacheMultipleTypes(t *testing.T) {
	t.Parallel()

	c1 := NewHeavy[string, int]()
	c1.Set("key1", 42)
	val1, ok := c1.Get("key1")
	if !ok || val1 != 42 {
		t.Errorf("Cache with string key and int value failed, got %v, %v", val1, ok)
	}

	c2 := NewHeavy[int, string]()
	c2.Set(1, "value1")
	val2, ok := c2.Get(1)
	if !ok || val2 != "value1" {
		t.Errorf("Cache with int key and string value failed, got %v, %v", val2, ok)
	}

	type Person struct {
		Name string
		Age  int
	}
	c3 := NewHeavy[string, Person]()
	person := Person{Name: "John", Age: 30}
	c3.Set("person1", person)
	val3, ok := c3.Get("person1")
	if !ok || val3.Name != "John" || val3.Age != 30 {
		t.Errorf("Cache with struct value failed, got %+v, %v", val3, ok)
	}
}

func TestHeavyCacheNilValueHandling(t *testing.T) {
	t.Parallel()

	c := NewHeavy[string, *int]()

	c.Set("nil-pointer", nil)

	val, ok := c.Get("nil-pointer")
	if !ok {
		t.Error("Get should return true for keys with nil values")
	}
	if val != nil {
		t.Errorf("Get should return nil for keys with nil values, got %v", val)
	}

	var zero int = 0
	c.Set("zero-pointer", &zero)
	val, ok = c.Get("zero-pointer")
	if !ok {
		t.Error("Get should return true for keys with zero value pointers")
	}
	if val == nil || *val != 0 {
		t.Errorf("Get should return zero pointer, got %v", val)
	}
}

func TestHeavyCacheEdgeCases(t *testing.T) {
	t.Parallel()

	c1 := NewHeavy[string, int]()
	c1.Set("", 42)
	val, ok := c1.Get("")
	if !ok || val != 42 {
		t.Errorf("Cache should handle empty string keys, got %v, %v", val, ok)
	}

	c2 := NewHeavy[int, string]()
	c2.Set(0, "zero-key")
	val2, ok := c2.Get(0)
	if !ok || val2 != "zero-key" {
		t.Errorf("Cache should handle zero value keys, got %v, %v", val2, ok)
	}
}

// 並行テスト

func TestHeavyCacheConcurrency(t *testing.T) {
	t.Parallel()

	c := NewHeavy[int, int]()
	var wg sync.WaitGroup

	for i := range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Set(i, i*10)
		}()
	}
	wg.Wait()

	start := make(chan struct{})
	for i := range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, _ = c.Get(i)
		}()
	}
	for i := 0; i < 50; i += 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			c.Delete(i)
		}()
	}
	close(start)
	wg.Wait()

	for i := 0; i < 50; i += 2 {
		_, ok := c.Get(i)
		if ok {
			t.Errorf("Key %d should have been deleted", i)
		}
	}

	for i := 1; i < 50; i += 2 {
		val, ok := c.Get(i)
		if !ok {
			t.Errorf("Key %d should exist", i)
		} else if val != i*10 {
			t.Errorf("Key %d should have value %d, got %d", i, i*10, val)
		}
	}
}

func TestHeavyCacheConcurrentClear(t *testing.T) {
	t.Parallel()

	c := NewHeavy[int, int]()
	var wg sync.WaitGroup

	for i := range 100 {
		c.Set(i, i)
	}

	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Clear()
		}()
	}
	for i := range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Set(i, i*100)
		}()
	}
	for i := range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = c.Get(i)
		}()
	}

	wg.Wait()
}

func TestHeavyCacheSetIfAbsentConcurrency(t *testing.T) {
	t.Parallel()

	c := NewHeavy[string, int]()
	var wg sync.WaitGroup
	results := make([]bool, 100)

	for i := range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i] = c.SetIfAbsent("key", i)
		}()
	}
	wg.Wait()

	count := 0
	for _, ok := range results {
		if ok {
			count++
		}
	}
	if count != 1 {
		t.Errorf("SetIfAbsent should succeed exactly once for the same key, got %d successes", count)
	}
}

func TestHeavyCacheGetOrSetConcurrency(t *testing.T) {
	t.Parallel()

	c := NewHeavy[string, int]()
	var wg sync.WaitGroup

	for i := range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = c.GetOrSet("key", i)
		}()
	}
	wg.Wait()

	val, ok := c.Get("key")
	if !ok {
		t.Error("GetOrSet should have set a value")
	}
	if val < 0 || val >= 100 {
		t.Errorf("GetOrSet result out of expected range: %d", val)
	}
}

func TestHeavyCacheDrainConcurrency(t *testing.T) {
	t.Parallel()

	c := NewHeavy[int, int]()
	var wg sync.WaitGroup

	for i := range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Set(i, i*10)
		}()
	}
	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Drain()
		}()
	}
	for i := range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = c.Get(i)
		}()
	}

	wg.Wait()
}

func TestHeavyCacheInterface(t *testing.T) {
	t.Parallel()

	var cache CacheInterface[string, int] = NewHeavy[string, int]()

	cache.Set("key1", 42)
	val, ok := cache.Get("key1")
	if !ok || val != 42 {
		t.Errorf("Interface Get failed, got %v, %v", val, ok)
	}

	cache.Delete("key1")
	_, ok = cache.Get("key1")
	if ok {
		t.Error("Interface Delete should remove keys")
	}

	cache.Set("key2", 100)
	cache.Clear()
	_, ok = cache.Get("key2")
	if ok {
		t.Error("Interface Clear should remove all keys")
	}

	cache.Set("key3", 300)
	cache.Set("key4", 400)
	if n := cache.Len(); n != 2 {
		t.Errorf("Interface Len should return 2, got %d", n)
	}
	if !cache.Has("key3") {
		t.Error("Interface Has should return true for existing key")
	}
	if len(cache.Keys()) != 2 {
		t.Errorf("Interface Keys should return 2 keys, got %d", len(cache.Keys()))
	}
	if len(cache.Values()) != 2 {
		t.Errorf("Interface Values should return 2 values, got %d", len(cache.Values()))
	}
	if ok := cache.SetIfAbsent("key3", 999); ok {
		t.Error("Interface SetIfAbsent should return false for existing key")
	}
	if ok := cache.SetIfAbsent("key5", 500); !ok {
		t.Error("Interface SetIfAbsent should return true for new key")
	}
	val5 := cache.GetOrSet("key5", 999)
	if val5 != 500 {
		t.Errorf("Interface GetOrSet should return existing value 500, got %d", val5)
	}
	allItems := cache.Snapshot()
	if len(allItems) != 3 {
		t.Errorf("Interface Snapshot should return 3 items, got %d", len(allItems))
	}
	items := cache.Drain()
	if len(items) != 3 {
		t.Errorf("Interface Drain should return all items, expected 3, got %d", len(items))
	}
	_, ok = cache.Get("key3")
	if ok {
		t.Error("Interface Drain should empty the cache")
	}
}

// 大量データテスト

func TestHeavyCacheLargeScale(t *testing.T) {
	t.Parallel()

	const n = 100_000
	c := NewHeavy[int, int]()

	for i := range n {
		c.Set(i, i*10)
	}
	if c.Len() != n {
		t.Errorf("Len should return %d, got %d", n, c.Len())
	}

	for i := range n {
		val, ok := c.Get(i)
		if !ok {
			t.Fatalf("Key %d should exist", i)
		}
		if val != i*10 {
			t.Fatalf("Key %d should have value %d, got %d", i, i*10, val)
		}
	}

	if len(c.Keys()) != n {
		t.Errorf("Keys should return %d keys, got %d", n, len(c.Keys()))
	}
	if len(c.Values()) != n {
		t.Errorf("Values should return %d values, got %d", n, len(c.Values()))
	}
	if len(c.Snapshot()) != n {
		t.Errorf("Snapshot should return %d items, got %d", n, len(c.Snapshot()))
	}

	drained := c.Drain()
	if len(drained) != n {
		t.Errorf("Drain should return %d items, got %d", n, len(drained))
	}
	if c.Len() != 0 {
		t.Errorf("Len should be 0 after Drain, got %d", c.Len())
	}
}

func TestHeavyCacheLargeScaleConcurrency(t *testing.T) {
	t.Parallel()

	const n = 100_000
	c := NewHeavy[int, int]()
	var wg sync.WaitGroup

	start := make(chan struct{})
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			c.Set(i, i*10)
		}()
	}
	close(start)
	wg.Wait()

	if c.Len() != n {
		t.Errorf("Len should return %d, got %d", n, c.Len())
	}

	start2 := make(chan struct{})
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start2
			switch i % 4 {
			case 0:
				c.Set(i, i*100)
			case 1:
				c.Get(i)
			case 2:
				c.Has(i)
			case 3:
				c.Delete(i)
			}
		}()
	}
	close(start2)
	wg.Wait()
}

// ベンチマーク

func BenchmarkHeavyCacheSet(b *testing.B) {
	c := NewHeavy[string, int]()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Set("key", i)
	}
}

func BenchmarkHeavyCacheGet(b *testing.B) {
	c := NewHeavy[string, int]()
	c.Set("key", 42)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Get("key")
	}
}

func BenchmarkHeavyCacheSetGet(b *testing.B) {
	c := NewHeavy[int, int]()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Set(i, i)
		c.Get(i)
	}
}

func BenchmarkHeavyCacheConcurrentReadWrite(b *testing.B) {
	c := NewHeavy[int, int]()
	for i := range 1000 {
		c.Set(i, i)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%10 == 0 {
				c.Set(i%1000, i)
			} else {
				c.Get(i % 1000)
			}
			i++
		}
	})
}

func BenchmarkHeavyCacheConcurrentRead(b *testing.B) {
	c := NewHeavy[int, int]()
	for i := range 1000 {
		c.Set(i, i)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			c.Get(i % 1000)
			i++
		}
	})
}

func BenchmarkHeavyCacheSetIfAbsent(b *testing.B) {
	c := NewHeavy[int, int]()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.SetIfAbsent(i, i)
	}
}

func BenchmarkHeavyCacheGetOrSet(b *testing.B) {
	c := NewHeavy[int, int]()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.GetOrSet(i%100, i)
	}
}

func BenchmarkHeavyCacheKeys(b *testing.B) {
	for _, size := range []int{100, 1_000, 10_000} {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			c := NewHeavy[int, int]()
			for i := range size {
				c.Set(i, i)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				c.Keys()
			}
		})
	}
}

func BenchmarkHeavyCacheSnapshot(b *testing.B) {
	for _, size := range []int{100, 1_000, 10_000} {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			c := NewHeavy[int, int]()
			for i := range size {
				c.Set(i, i)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				c.Snapshot()
			}
		})
	}
}

func TestHeavyCacheDeleteFunc(t *testing.T) {
	t.Parallel()

	c := NewHeavy[string, int]()
	c.Set("a", 1)
	c.Set("b", -2)
	c.Set("c", 3)
	c.Set("d", -4)

	c.DeleteFunc(func(k string, v int) bool {
		return v <= 0
	})

	if c.Len() != 2 {
		t.Errorf("Len should be 2 after DeleteFunc, got %d", c.Len())
	}
	if v, ok := c.Get("a"); !ok || v != 1 {
		t.Errorf("key 'a' should remain with value 1, got %v, %v", v, ok)
	}
	if v, ok := c.Get("c"); !ok || v != 3 {
		t.Errorf("key 'c' should remain with value 3, got %v, %v", v, ok)
	}
	if _, ok := c.Get("b"); ok {
		t.Error("key 'b' should have been deleted")
	}
	if _, ok := c.Get("d"); ok {
		t.Error("key 'd' should have been deleted")
	}

	// 全件削除
	c2 := NewHeavy[string, int]()
	c2.Set("x", 1)
	c2.Set("y", 2)
	c2.DeleteFunc(func(k string, v int) bool { return true })
	if c2.Len() != 0 {
		t.Errorf("DeleteFunc with always-true should delete all, got len %d", c2.Len())
	}
}

func TestHeavyCacheDeleteFuncConcurrency(t *testing.T) {
	t.Parallel()

	c := NewHeavy[int, int]()
	var wg sync.WaitGroup

	for i := range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Set(i, i)
		}()
	}
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.DeleteFunc(func(k, v int) bool { return k%2 == 0 })
		}()
	}

	wg.Wait()
}

func TestHeavyCachePop(t *testing.T) {
	t.Parallel()

	c := NewHeavy[string, int]()
	c.Set("key1", 42)
	c.Set("key2", 100)

	val, ok := c.Pop("key1")
	if !ok || val != 42 {
		t.Errorf("Pop should return (42, true), got (%v, %v)", val, ok)
	}
	if c.Has("key1") {
		t.Error("Pop should remove the key from cache")
	}

	// 存在しないキー
	val, ok = c.Pop("not-exists")
	if ok {
		t.Error("Pop should return false for non-existent key")
	}
	if val != 0 {
		t.Errorf("Pop should return zero value for non-existent key, got %v", val)
	}
}

func TestHeavyCachePopConcurrency(t *testing.T) {
	t.Parallel()

	c := NewHeavy[int, int]()
	for i := range 100 {
		c.Set(i, i*10)
	}

	var wg sync.WaitGroup
	results := make([]bool, 100)
	for i := range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, results[i] = c.Pop(i)
		}()
	}
	wg.Wait()

	for i := range 100 {
		if !results[i] {
			t.Errorf("Pop(%d) should have succeeded", i)
		}
	}
	if c.Len() != 0 {
		t.Errorf("Cache should be empty after popping all keys, got len %d", c.Len())
	}
}

func TestHeavyCacheSetAll(t *testing.T) {
	t.Parallel()

	c := NewHeavy[string, int]()
	c.Set("existing", 999)

	c.SetAll(map[string]int{
		"key1": 1,
		"key2": 2,
		"key3": 3,
	})

	if c.Len() != 4 {
		t.Errorf("Len should be 4, got %d", c.Len())
	}
	for _, tc := range []struct {
		k string
		v int
	}{{"existing", 999}, {"key1", 1}, {"key2", 2}, {"key3", 3}} {
		val, ok := c.Get(tc.k)
		if !ok || val != tc.v {
			t.Errorf("Get(%q) = (%v, %v), want (%v, true)", tc.k, val, ok, tc.v)
		}
	}

	// 上書き
	c.SetAll(map[string]int{"existing": 0})
	val, _ := c.Get("existing")
	if val != 0 {
		t.Errorf("SetAll should overwrite existing keys, got %v", val)
	}
}

func TestHeavyCacheSetAllConcurrency(t *testing.T) {
	t.Parallel()

	c := NewHeavy[int, int]()
	var wg sync.WaitGroup

	for i := range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.SetAll(map[int]int{i: i * 10, i + 1000: i * 100})
		}()
	}

	wg.Wait()
}

func TestHeavyCacheGetAll(t *testing.T) {
	t.Parallel()

	c := NewHeavy[string, int]()
	c.Set("key1", 1)
	c.Set("key3", 3)

	result := c.GetAll("key1", "key2", "key3")
	if len(result) != 2 {
		t.Errorf("GetAll should return 2 items, got %d", len(result))
	}
	if result["key1"] != 1 {
		t.Errorf("GetAll should contain key1=1, got %v", result["key1"])
	}
	if result["key3"] != 3 {
		t.Errorf("GetAll should contain key3=3, got %v", result["key3"])
	}
	if _, exists := result["key2"]; exists {
		t.Error("GetAll should not contain non-existent keys")
	}

	// キーなし
	empty := c.GetAll()
	if len(empty) != 0 {
		t.Errorf("GetAll with no keys should return empty map, got %v", empty)
	}
}

func TestHeavyCacheGetAllConcurrency(t *testing.T) {
	t.Parallel()

	c := NewHeavy[int, int]()
	for i := range 100 {
		c.Set(i, i*10)
	}

	var wg sync.WaitGroup
	for i := range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = c.GetAll(i, i+1, i+2)
		}()
	}
	for i := range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Set(i, i*100)
		}()
	}

	wg.Wait()
}

func TestHeavyCacheGetOrSetFunc(t *testing.T) {
	t.Parallel()

	c := NewHeavy[string, int]()

	callCount := 0
	val := c.GetOrSetFunc("key1", func() int {
		callCount++
		return 42
	})
	if val != 42 {
		t.Errorf("GetOrSetFunc should return 42, got %v", val)
	}
	if callCount != 1 {
		t.Errorf("fn should be called once, got %d", callCount)
	}

	// 既に存在するキーではfnは呼ばれない
	val = c.GetOrSetFunc("key1", func() int {
		callCount++
		return 999
	})
	if val != 42 {
		t.Errorf("GetOrSetFunc should return existing value 42, got %v", val)
	}
	if callCount != 1 {
		t.Errorf("fn should not be called for existing key, got %d calls", callCount)
	}
}

func TestHeavyCacheGetOrSetFuncConcurrency(t *testing.T) {
	t.Parallel()

	c := NewHeavy[string, int]()
	var wg sync.WaitGroup

	for i := range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = c.GetOrSetFunc("key", func() int { return i })
		}()
	}
	wg.Wait()

	val, ok := c.Get("key")
	if !ok {
		t.Error("GetOrSetFunc should have set a value")
	}
	if val < 0 || val >= 100 {
		t.Errorf("GetOrSetFunc result out of expected range: %d", val)
	}
}

func TestHeavyCacheCount(t *testing.T) {
	t.Parallel()

	c := NewHeavy[string, int]()
	c.Set("a", 1)
	c.Set("b", -2)
	c.Set("c", 3)
	c.Set("d", -4)

	n := c.Count(func(k string, v int) bool { return v > 0 })
	if n != 2 {
		t.Errorf("Count of positive values should be 2, got %d", n)
	}

	n = c.Count(func(k string, v int) bool { return true })
	if n != 4 {
		t.Errorf("Count with always-true should be 4, got %d", n)
	}
}

func TestHeavyCacheUpdate(t *testing.T) {
	t.Parallel()

	c := NewHeavy[string, int]()
	c.Set("counter", 10)

	ok := c.Update("counter", func(v int) int { return v + 1 })
	if !ok {
		t.Error("Update should return true for existing key")
	}
	val, _ := c.Get("counter")
	if val != 11 {
		t.Errorf("Update should increment value, expected 11, got %d", val)
	}

	ok = c.Update("missing", func(v int) int { return v + 1 })
	if ok {
		t.Error("Update should return false for non-existent key")
	}
	if c.Has("missing") {
		t.Error("Update should not create a new key")
	}
}

func TestHeavyCacheUpdateConcurrency(t *testing.T) {
	t.Parallel()

	c := NewHeavy[string, int]()
	c.Set("counter", 0)
	var wg sync.WaitGroup

	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Update("counter", func(v int) int { return v + 1 })
		}()
	}
	wg.Wait()

	val, _ := c.Get("counter")
	if val != 100 {
		t.Errorf("Counter should be 100 after 100 increments, got %d", val)
	}
}

func TestHeavyCacheSwap(t *testing.T) {
	t.Parallel()

	c := NewHeavy[string, int]()
	c.Set("key1", 42)

	old, ok := c.Swap("key1", 100)
	if !ok {
		t.Error("Swap should return true for existing key")
	}
	if old != 42 {
		t.Errorf("Swap should return old value 42, got %d", old)
	}
	val, _ := c.Get("key1")
	if val != 100 {
		t.Errorf("Swap should set new value 100, got %d", val)
	}

	old, ok = c.Swap("new-key", 200)
	if ok {
		t.Error("Swap should return false for non-existent key")
	}
	if old != 0 {
		t.Errorf("Swap should return zero value for non-existent key, got %d", old)
	}
	val, exists := c.Get("new-key")
	if !exists || val != 200 {
		t.Errorf("Swap should create new key with value 200, got %v, %v", val, exists)
	}
}

func TestHeavyCacheSwapConcurrency(t *testing.T) {
	t.Parallel()

	c := NewHeavy[string, int]()
	c.Set("key", 0)
	var wg sync.WaitGroup

	for i := range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Swap("key", i)
		}()
	}
	wg.Wait()

	_, ok := c.Get("key")
	if !ok {
		t.Error("key should exist after concurrent swaps")
	}
}

func TestHeavyCacheCompareAndSwap(t *testing.T) {
	t.Parallel()

	eq := func(a, b int) bool { return a == b }
	c := NewHeavy[string, int]()
	c.Set("key1", 42)

	ok := c.CompareAndSwap("key1", 42, 100, eq)
	if !ok {
		t.Error("CompareAndSwap should return true when values match")
	}
	val, _ := c.Get("key1")
	if val != 100 {
		t.Errorf("CompareAndSwap should set new value 100, got %d", val)
	}

	ok = c.CompareAndSwap("key1", 42, 200, eq)
	if ok {
		t.Error("CompareAndSwap should return false when values don't match")
	}
	val, _ = c.Get("key1")
	if val != 100 {
		t.Errorf("CompareAndSwap should not change value, expected 100, got %d", val)
	}

	ok = c.CompareAndSwap("missing", 0, 1, eq)
	if ok {
		t.Error("CompareAndSwap should return false for non-existent key")
	}
}

func TestHeavyCacheCompareAndSwapConcurrency(t *testing.T) {
	t.Parallel()

	eq := func(a, b int) bool { return a == b }
	c := NewHeavy[string, int]()
	c.Set("key", 0)
	var wg sync.WaitGroup

	results := make([]bool, 100)
	for i := range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i] = c.CompareAndSwap("key", 0, 1, eq)
		}()
	}
	wg.Wait()

	count := 0
	for _, ok := range results {
		if ok {
			count++
		}
	}
	if count != 1 {
		t.Errorf("CompareAndSwap should succeed exactly once, got %d", count)
	}
}

func TestHeavyCacheReplace(t *testing.T) {
	t.Parallel()

	c := NewHeavy[string, int]()

	ok := c.Replace("missing", 100)
	if ok {
		t.Error("Replace should return false for non-existent key")
	}
	if c.Has("missing") {
		t.Error("Replace should not create a new key")
	}

	c.Set("key1", 42)
	ok = c.Replace("key1", 100)
	if !ok {
		t.Error("Replace should return true for existing key")
	}
	val, _ := c.Get("key1")
	if val != 100 {
		t.Errorf("Replace should update value to 100, got %d", val)
	}

	c.Delete("key1")
	ok = c.Replace("key1", 200)
	if ok {
		t.Error("Replace should return false after Delete")
	}
}

func TestHeavyCacheReplaceConcurrency(t *testing.T) {
	t.Parallel()

	c := NewHeavy[int, int]()
	for i := range 100 {
		c.Set(i, i)
	}
	var wg sync.WaitGroup

	for i := range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Replace(i, i*100)
		}()
	}
	for i := range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Replace(i+200, 999)
		}()
	}
	wg.Wait()
}

func TestHeavyCacheMap(t *testing.T) {
	t.Parallel()

	c := NewHeavy[string, int]()
	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)

	c.Map(func(k string, v int) int { return v * 2 })

	expected := map[string]int{"a": 2, "b": 4, "c": 6}
	for k, want := range expected {
		got, ok := c.Get(k)
		if !ok || got != want {
			t.Errorf("Map: Get(%q) = (%d, %v), want (%d, true)", k, got, ok, want)
		}
	}

	if c.Len() != 3 {
		t.Errorf("Map should not change cache size, got %d", c.Len())
	}

	c2 := NewHeavy[string, int]()
	c2.Map(func(k string, v int) int { return v * 2 })
	if c2.Len() != 0 {
		t.Errorf("Map on empty cache should be safe, got len %d", c2.Len())
	}
}

func TestHeavyCacheMapConcurrency(t *testing.T) {
	t.Parallel()

	c := NewHeavy[int, int]()
	for i := range 100 {
		c.Set(i, i)
	}
	var wg sync.WaitGroup

	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Map(func(k, v int) int { return v + 1 })
		}()
	}
	for i := range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = c.Get(i)
		}()
	}
	wg.Wait()
}

func TestHeavyCacheDeleteAll(t *testing.T) {
	t.Parallel()

	c := NewHeavy[string, int]()
	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)
	c.Set("d", 4)

	c.DeleteAll("a", "c")
	if c.Len() != 2 {
		t.Errorf("Len should be 2 after DeleteAll, got %d", c.Len())
	}
	if c.Has("a") || c.Has("c") {
		t.Error("DeleteAll should remove specified keys")
	}
	if !c.Has("b") || !c.Has("d") {
		t.Error("DeleteAll should not remove unspecified keys")
	}

	// 存在しないキーを含んでも安全
	c.DeleteAll("b", "not-exists")
	if c.Len() != 1 {
		t.Errorf("Len should be 1, got %d", c.Len())
	}

	// 引数なしでも安全
	c.DeleteAll()
	if c.Len() != 1 {
		t.Errorf("DeleteAll with no args should not change cache, got len %d", c.Len())
	}
}

func TestHeavyCacheDeleteAllConcurrency(t *testing.T) {
	t.Parallel()

	c := NewHeavy[int, int]()
	for i := range 100 {
		c.Set(i, i)
	}
	var wg sync.WaitGroup

	for i := range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.DeleteAll(i*5, i*5+1, i*5+2, i*5+3, i*5+4)
		}()
	}
	for i := range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = c.Get(i)
		}()
	}
	wg.Wait()
}

func TestHeavyCacheFilter(t *testing.T) {
	t.Parallel()

	c := NewHeavy[string, int]()
	c.Set("a", 1)
	c.Set("b", -2)
	c.Set("c", 3)
	c.Set("d", -4)

	result := c.Filter(func(k string, v int) bool { return v > 0 })
	if len(result) != 2 {
		t.Errorf("Filter should return 2 items, got %d", len(result))
	}
	if result["a"] != 1 || result["c"] != 3 {
		t.Errorf("Filter should return positive values, got %v", result)
	}

	if c.Len() != 4 {
		t.Errorf("Filter should not modify cache, got len %d", c.Len())
	}

	result["a"] = 999
	val, _ := c.Get("a")
	if val != 1 {
		t.Errorf("Modifying Filter result should not affect cache, got %d", val)
	}

	empty := c.Filter(func(k string, v int) bool { return false })
	if len(empty) != 0 {
		t.Errorf("Filter with always-false should return empty map, got %v", empty)
	}
}

func TestHeavyCacheFilterConcurrency(t *testing.T) {
	t.Parallel()

	c := NewHeavy[int, int]()
	for i := range 100 {
		c.Set(i, i)
	}
	var wg sync.WaitGroup

	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = c.Filter(func(k, v int) bool { return v%2 == 0 })
		}()
	}
	for i := range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Set(i, i*100)
		}()
	}
	wg.Wait()
}

func TestHeavyCacheCompareAndDelete(t *testing.T) {
	t.Parallel()

	c := NewHeavy[string, int]()
	c.Set("key1", 42)

	ok := c.CompareAndDelete("key1", func(v int) bool { return v == 42 })
	if !ok {
		t.Error("CompareAndDelete should return true when condition matches")
	}
	if c.Has("key1") {
		t.Error("CompareAndDelete should remove the key")
	}

	c.Set("key2", 100)
	ok = c.CompareAndDelete("key2", func(v int) bool { return v == 42 })
	if ok {
		t.Error("CompareAndDelete should return false when condition doesn't match")
	}
	if !c.Has("key2") {
		t.Error("CompareAndDelete should not remove key when condition doesn't match")
	}

	ok = c.CompareAndDelete("missing", func(v int) bool { return true })
	if ok {
		t.Error("CompareAndDelete should return false for non-existent key")
	}
}

func TestHeavyCacheCompareAndDeleteConcurrency(t *testing.T) {
	t.Parallel()

	c := NewHeavy[int, int]()
	for i := range 100 {
		c.Set(i, i)
	}
	var wg sync.WaitGroup

	for i := range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.CompareAndDelete(i, func(v int) bool { return v%2 == 0 })
		}()
	}
	for i := range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = c.Get(i)
		}()
	}
	wg.Wait()
}

func BenchmarkHeavyCacheDrain(b *testing.B) {
	for _, size := range []int{100, 1_000, 10_000} {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			c := NewHeavy[int, int]()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for j := range size {
					c.Set(j, j)
				}
				c.Drain()
			}
		})
	}
}
