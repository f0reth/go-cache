package cache

import (
	"fmt"
	"sync"
	"testing"
)

func TestNew(t *testing.T) {
	t.Parallel()

	c := New[string, int]()
	if c == nil {
		t.Fatal("New should return a non-nil cache")
	}
	if c.items == nil {
		t.Fatal("New should initialize the items map")
	}
	if len(c.items) != 0 {
		t.Errorf("New cache should be empty, got size %d", len(c.items))
	}
}

func TestSetAndGet(t *testing.T) {
	t.Parallel()

	c := New[string, int]()

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

func TestDelete(t *testing.T) {
	t.Parallel()

	c := New[string, int]()

	// 値を設定
	c.Set("key1", 42)
	c.Set("key2", 100)

	// 値を削除
	c.Delete("key1")

	// 削除されたキーにアクセス
	val, ok := c.Get("key1")
	if ok {
		t.Error("Get should return false for deleted keys")
	}
	if val != 0 {
		t.Errorf("Get should return zero value for deleted keys, got %v", val)
	}

	// 削除されていないキーは残っているべき
	val, ok = c.Get("key2")
	if !ok {
		t.Error("Delete should not affect other keys")
	}
	if val != 100 {
		t.Errorf("Get should return the correct value for non-deleted keys, expected 100, got %v", val)
	}

	// 存在しないキーを削除してもエラーにならないことを確認
	c.Delete("not-exists")
	// 特にアサートはない - パニックがなければテストは通過
}

func TestClear(t *testing.T) {
	t.Parallel()

	c := New[string, int]()

	// いくつかの値を設定
	c.Set("key1", 42)
	c.Set("key2", 100)
	c.Set("key3", 200)

	// キャッシュをクリア
	c.Clear()

	// すべてのキーが削除されていることを確認
	for _, key := range []string{"key1", "key2", "key3"} {
		val, ok := c.Get(key)
		if ok {
			t.Errorf("Get should return false for key %s after Clear", key)
		}
		if val != 0 {
			t.Errorf("Get should return zero value for key %s after Clear, got %v", key, val)
		}
	}

	// 新しいキーを再設定できることを確認
	c.Set("new-key", 300)
	val, ok := c.Get("new-key")
	if !ok {
		t.Error("Set should work after Clear")
	}
	if val != 300 {
		t.Errorf("Get should return the correct value after Clear, expected 300, got %v", val)
	}
}

func TestMultipleTypes(t *testing.T) {
	t.Parallel()

	// 文字列キー・整数値
	c1 := New[string, int]()
	c1.Set("key1", 42)
	val1, ok := c1.Get("key1")
	if !ok || val1 != 42 {
		t.Errorf("Cache with string key and int value failed, got %v, %v", val1, ok)
	}

	// 整数キー・文字列値
	c2 := New[int, string]()
	c2.Set(1, "value1")
	val2, ok := c2.Get(1)
	if !ok || val2 != "value1" {
		t.Errorf("Cache with int key and string value failed, got %v, %v", val2, ok)
	}

	// 構造体をキーとして使用
	type ComplexKey struct {
		ID   int
		Name string
	}
	c3 := New[ComplexKey, float64]()
	key := ComplexKey{ID: 1, Name: "test"}
	c3.Set(key, 3.14)
	val3, ok := c3.Get(key)
	if !ok || val3 != 3.14 {
		t.Errorf("Cache with struct key and float value failed, got %v, %v", val3, ok)
	}

	// 構造体を値として使用
	type Person struct {
		Name  string
		Age   int
		Email string
	}
	c4 := New[string, Person]()
	person := Person{Name: "John", Age: 30, Email: "john@example.com"}
	c4.Set("person1", person)
	val4, ok := c4.Get("person1")
	if !ok || val4.Name != "John" || val4.Age != 30 || val4.Email != "john@example.com" {
		t.Errorf("Cache with string key and struct value failed, got %+v, %v", val4, ok)
	}
}

func TestConcurrency(t *testing.T) {
	t.Parallel()

	c := New[int, int]()
	var wg sync.WaitGroup

	// フェーズ1: 複数のゴルーチンで同時に書き込み
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			c.Set(id, id*10)
		}(i)
	}
	wg.Wait()

	// フェーズ2: バリアパターンで読み込みと削除を確実に同時実行
	start := make(chan struct{})
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			<-start
			_, _ = c.Get(id)
		}(i)
	}
	for i := 0; i < 50; i += 2 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			<-start
			c.Delete(id)
		}(i)
	}
	close(start) // 全ゴルーチンを一斉スタート
	wg.Wait()

	// 削除したキーが存在しないことを確認
	for i := 0; i < 50; i += 2 {
		_, ok := c.Get(i)
		if ok {
			t.Errorf("Key %d should have been deleted", i)
		}
	}

	// 残りのキーが正しい値を持つことを確認
	for i := 1; i < 50; i += 2 {
		val, ok := c.Get(i)
		if !ok {
			t.Errorf("Key %d should exist", i)
		} else if val != i*10 {
			t.Errorf("Key %d should have value %d, got %d", i, i*10, val)
		}
	}
}

func TestConcurrentClear(t *testing.T) {
	t.Parallel()

	c := New[int, int]()
	var wg sync.WaitGroup

	// キャッシュに値を追加
	for i := 0; i < 100; i++ {
		c.Set(i, i)
	}

	// いくつかのゴルーチンでクリア操作
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Clear()
		}()
	}

	// 同時に追加操作
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			c.Set(id, id*100)
		}(i)
	}

	// 同時に読み込み操作
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_, _ = c.Get(id)
		}(i)
	}

	// すべてのゴルーチンの完了を待つ
	wg.Wait()

	// 最終的な状態をチェックする必要はない
	// このテストは主にデッドロックや競合状態がないことを確認する
}

func TestNilValueHandling(t *testing.T) {
	t.Parallel()

	// ポインタ型の値を使用
	c := New[string, *int]()

	// nilポインタを格納
	c.Set("nil-pointer", nil)

	// 格納したnilを取得
	val, ok := c.Get("nil-pointer")
	if !ok {
		t.Error("Get should return true for keys with nil values")
	}
	if val != nil {
		t.Errorf("Get should return nil for keys with nil values, got %v", val)
	}

	// ゼロ値と未設定キーの区別
	var zero int = 0
	c.Set("zero-pointer", &zero)

	val, ok = c.Get("zero-pointer")
	if !ok {
		t.Error("Get should return true for keys with zero value pointers")
	}
	if val == nil {
		t.Error("Get should not return nil for keys with zero value pointers")
	} else if *val != 0 {
		t.Errorf("Get should return zero for keys with zero value pointers, got %v", *val)
	}
}

func TestEdgeCases(t *testing.T) {
	t.Parallel()

	// 空の文字列をキーとして使用
	c1 := New[string, int]()
	c1.Set("", 42)
	val, ok := c1.Get("")
	if !ok || val != 42 {
		t.Errorf("Cache should handle empty string keys, got %v, %v", val, ok)
	}

	// マップをクリアした後に再利用
	c1.Clear()
	c1.Set("key1", 100)
	val, ok = c1.Get("key1")
	if !ok || val != 100 {
		t.Errorf("Cache should be reusable after Clear, got %v, %v", val, ok)
	}

	// ゼロ値のキー
	c2 := New[int, string]()
	c2.Set(0, "zero-key")
	val2, ok := c2.Get(0)
	if !ok || val2 != "zero-key" {
		t.Errorf("Cache should handle zero value keys, got %v, %v", val2, ok)
	}
}

func TestDrain(t *testing.T) {
	t.Parallel()

	// 基本動作: 複数の値を取り出し、キャッシュが空になることを確認
	c := New[string, int]()
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

	// キャッシュが空になっていることを確認
	for _, key := range []string{"key1", "key2", "key3"} {
		_, ok := c.Get(key)
		if ok {
			t.Errorf("Cache should be empty after Drain, but key %s still exists", key)
		}
	}

	// Drain後にキャッシュが再利用できることを確認
	c.Set("new-key", 100)
	val, ok := c.Get("new-key")
	if !ok || val != 100 {
		t.Errorf("Cache should be reusable after Drain, got %v, %v", val, ok)
	}
}

func TestDrainEmpty(t *testing.T) {
	t.Parallel()

	// 空のキャッシュをDrainしても空のマップが返ることを確認
	c := New[string, int]()
	items := c.Drain()

	if len(items) != 0 {
		t.Errorf("Drain on empty cache should return empty map, got %d items", len(items))
	}
}

func TestDrainTwice(t *testing.T) {
	t.Parallel()

	// 2回連続でDrainした場合、2回目は空のマップが返ることを確認
	c := New[string, int]()
	c.Set("key1", 1)

	first := c.Drain()
	if len(first) != 1 {
		t.Errorf("First Drain should return 1 item, got %d", len(first))
	}

	second := c.Drain()
	if len(second) != 0 {
		t.Errorf("Second Drain should return empty map, got %d items", len(second))
	}
}

func TestDrainIsolation(t *testing.T) {
	t.Parallel()

	// Drainで取得したマップを変更してもキャッシュに影響しないことを確認
	c := New[string, int]()
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

func TestDrainConcurrency(t *testing.T) {
	t.Parallel()

	// 並行してDrain・Set・Getを実行してもデッドロックや競合が起きないことを確認
	c := New[int, int]()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			c.Set(id, id*10)
		}(i)
	}

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Drain()
		}()
	}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_, _ = c.Get(id)
		}(i)
	}

	wg.Wait()
}

func TestHas(t *testing.T) {
	t.Parallel()

	c := New[string, int]()

	// 存在しないキーはfalseを返す
	if c.Has("not-exists") {
		t.Error("Has should return false for non-existent keys")
	}

	// 存在するキーはtrueを返す
	c.Set("key1", 42)
	if !c.Has("key1") {
		t.Error("Has should return true for existing keys")
	}

	// 削除後はfalseを返す
	c.Delete("key1")
	if c.Has("key1") {
		t.Error("Has should return false after Delete")
	}

	// nilポインタを格納したキーもtrueを返す
	cp := New[string, *int]()
	cp.Set("nil-key", nil)
	if !cp.Has("nil-key") {
		t.Error("Has should return true for keys with nil values")
	}

	// Clear後はfalseを返す
	c.Set("key2", 100)
	c.Clear()
	if c.Has("key2") {
		t.Error("Has should return false after Clear")
	}
}

func TestHasConcurrency(t *testing.T) {
	t.Parallel()

	c := New[int, int]()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			c.Set(id, id*10)
		}(i)
	}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_ = c.Has(id)
		}(i)
	}

	wg.Wait()
}

func TestKeys(t *testing.T) {
	t.Parallel()

	c := New[string, int]()

	// 空のキャッシュは空スライスを返す
	keys := c.Keys()
	if len(keys) != 0 {
		t.Errorf("Keys should return empty slice for empty cache, got %v", keys)
	}

	// 追加したキーが含まれる
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

	// 削除後はそのキーが含まれない
	c.Delete("key1")
	keys = c.Keys()
	if len(keys) != 2 {
		t.Errorf("Keys should return 2 keys after Delete, got %d", len(keys))
	}
	for _, k := range keys {
		if k == "key1" {
			t.Error("Keys should not contain deleted key")
		}
	}

	// Clear後は空スライスを返す
	c.Clear()
	keys = c.Keys()
	if len(keys) != 0 {
		t.Errorf("Keys should return empty slice after Clear, got %v", keys)
	}
}

func TestKeysConcurrency(t *testing.T) {
	t.Parallel()

	c := New[int, int]()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			c.Set(id, id*10)
		}(i)
	}
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = c.Keys()
		}()
	}

	wg.Wait()
}

func TestValues(t *testing.T) {
	t.Parallel()

	c := New[string, int]()

	// 空のキャッシュは空スライスを返す
	values := c.Values()
	if len(values) != 0 {
		t.Errorf("Values should return empty slice for empty cache, got %v", values)
	}

	// 追加した値が含まれる
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

	// 削除後はその値が含まれない
	c.Delete("key1")
	values = c.Values()
	if len(values) != 2 {
		t.Errorf("Values should return 2 values after Delete, got %d", len(values))
	}
	for _, v := range values {
		if v == 1 {
			t.Error("Values should not contain value of deleted key")
		}
	}

	// Clear後は空スライスを返す
	c.Clear()
	values = c.Values()
	if len(values) != 0 {
		t.Errorf("Values should return empty slice after Clear, got %v", values)
	}
}

func TestValuesConcurrency(t *testing.T) {
	t.Parallel()

	c := New[int, int]()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			c.Set(id, id*10)
		}(i)
	}
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = c.Values()
		}()
	}

	wg.Wait()
}

func TestSetIfAbsent(t *testing.T) {
	t.Parallel()

	c := New[string, int]()

	// 存在しないキーはセットされtrueを返す
	if ok := c.SetIfAbsent("key1", 42); !ok {
		t.Error("SetIfAbsent should return true when key does not exist")
	}
	val, _ := c.Get("key1")
	if val != 42 {
		t.Errorf("SetIfAbsent should set the value, expected 42, got %v", val)
	}

	// 既に存在するキーは上書きされずfalseを返す
	if ok := c.SetIfAbsent("key1", 999); ok {
		t.Error("SetIfAbsent should return false when key already exists")
	}
	val, _ = c.Get("key1")
	if val != 42 {
		t.Errorf("SetIfAbsent should not overwrite existing value, expected 42, got %v", val)
	}

	// 削除後は再セットできる
	c.Delete("key1")
	if ok := c.SetIfAbsent("key1", 100); !ok {
		t.Error("SetIfAbsent should return true after Delete")
	}
	val, _ = c.Get("key1")
	if val != 100 {
		t.Errorf("SetIfAbsent should set new value after Delete, expected 100, got %v", val)
	}
}

func TestSetIfAbsentConcurrency(t *testing.T) {
	t.Parallel()

	c := New[string, int]()
	var wg sync.WaitGroup
	results := make([]bool, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			results[id] = c.SetIfAbsent("key", id)
		}(i)
	}
	wg.Wait()

	// 同一キーに対してtrueを返したゴルーチンはちょうど1つ
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

func TestGetOrSet(t *testing.T) {
	t.Parallel()

	c := New[string, int]()

	// 存在しないキーはvalueをセットして返す
	val := c.GetOrSet("key1", 42)
	if val != 42 {
		t.Errorf("GetOrSet should return the set value, expected 42, got %v", val)
	}
	stored, ok := c.Get("key1")
	if !ok || stored != 42 {
		t.Errorf("GetOrSet should persist the value in cache, expected 42, got %v", stored)
	}

	// 既に存在するキーは既存の値を返し、上書きしない
	val = c.GetOrSet("key1", 999)
	if val != 42 {
		t.Errorf("GetOrSet should return existing value, expected 42, got %v", val)
	}
	stored, _ = c.Get("key1")
	if stored != 42 {
		t.Errorf("GetOrSet should not overwrite existing value, expected 42, got %v", stored)
	}
}

func TestGetOrSetConcurrency(t *testing.T) {
	t.Parallel()

	c := New[string, int]()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_ = c.GetOrSet("key", id)
		}(i)
	}
	wg.Wait()

	// 結果は必ずいずれか1つのgoroutineがセットした値
	val, ok := c.Get("key")
	if !ok {
		t.Error("GetOrSet should have set a value")
	}
	if val < 0 || val >= 100 {
		t.Errorf("GetOrSet result out of expected range: %d", val)
	}
}

func TestSnapshot(t *testing.T) {
	t.Parallel()

	c := New[string, int]()

	// 空のキャッシュは空マップを返す
	items := c.Snapshot()
	if len(items) != 0 {
		t.Errorf("Snapshot should return empty map for empty cache, got %v", items)
	}

	// 追加したアイテムがすべて含まれる
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

	// キャッシュは変更されない（Drainと違い）
	if c.Len() != 3 {
		t.Errorf("Snapshot should not remove items from cache, got len %d", c.Len())
	}

	// 返されたマップを変更してもキャッシュに影響しない
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

func TestSnapshotConcurrency(t *testing.T) {
	t.Parallel()

	c := New[int, int]()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			c.Set(id, id*10)
		}(i)
	}
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = c.Snapshot()
		}()
	}

	wg.Wait()
}

func TestRange(t *testing.T) {
	t.Parallel()

	c := New[string, int]()

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

func TestRangeConcurrency(t *testing.T) {
	t.Parallel()

	c := New[int, int]()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			c.Set(id, id*10)
		}(i)
	}
	for i := 0; i < 50; i++ {
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

func TestLen(t *testing.T) {
	t.Parallel()

	c := New[string, int]()

	// 空のキャッシュは0を返す
	if n := c.Len(); n != 0 {
		t.Errorf("Len should return 0 for empty cache, got %d", n)
	}

	// アイテムを追加するたびに増加する
	c.Set("key1", 1)
	if n := c.Len(); n != 1 {
		t.Errorf("Len should return 1 after one Set, got %d", n)
	}

	c.Set("key2", 2)
	c.Set("key3", 3)
	if n := c.Len(); n != 3 {
		t.Errorf("Len should return 3 after three Sets, got %d", n)
	}

	// 同じキーを上書きしてもアイテム数は変わらない
	c.Set("key1", 100)
	if n := c.Len(); n != 3 {
		t.Errorf("Len should return 3 after overwriting a key, got %d", n)
	}

	// 削除するとアイテム数が減る
	c.Delete("key1")
	if n := c.Len(); n != 2 {
		t.Errorf("Len should return 2 after Delete, got %d", n)
	}

	// Clearするとアイテム数が0になる
	c.Clear()
	if n := c.Len(); n != 0 {
		t.Errorf("Len should return 0 after Clear, got %d", n)
	}

	// Drainするとアイテム数が0になる
	c.Set("key1", 1)
	c.Set("key2", 2)
	c.Drain()
	if n := c.Len(); n != 0 {
		t.Errorf("Len should return 0 after Drain, got %d", n)
	}
}

func TestLenConcurrency(t *testing.T) {
	t.Parallel()

	c := New[int, int]()
	var wg sync.WaitGroup

	// 並行してSetとLenを実行してもデッドロックや競合が起きないことを確認
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			c.Set(id, id*10)
		}(i)
	}
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = c.Len()
		}()
	}

	wg.Wait()

	if n := c.Len(); n != 100 {
		t.Errorf("Len should return 100 after 100 Sets, got %d", n)
	}
}

func TestInterface(t *testing.T) {
	t.Parallel()

	var cache CacheInterface[string, int] = New[string, int]()

	// インターフェースメソッドを使用
	cache.Set("key1", 42)
	val, ok := cache.Get("key1")
	if !ok {
		t.Error("Interface Get should return true for existing keys")
	}
	if val != 42 {
		t.Errorf("Interface Get should return the correct value, expected 42, got %v", val)
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
	if cache.Has("not-exists") {
		t.Error("Interface Has should return false for non-existent key")
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
	if cache.Len() != 3 {
		t.Error("Interface Snapshot should not remove items from cache")
	}
	items := cache.Drain()
	if len(items) != 3 {
		t.Errorf("Interface Drain should return all items, expected 3, got %d", len(items))
	}
	if items["key3"] != 300 || items["key4"] != 400 || items["key5"] != 500 {
		t.Errorf("Interface Drain should return correct values, got %v", items)
	}
	_, ok = cache.Get("key3")
	if ok {
		t.Error("Interface Drain should empty the cache")
	}
}

// ベンチマークテスト

func BenchmarkSet(b *testing.B) {
	c := New[string, int]()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Set("key", i)
	}
}

func BenchmarkGet(b *testing.B) {
	c := New[string, int]()
	c.Set("key", 42)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Get("key")
	}
}

func BenchmarkSetGet(b *testing.B) {
	c := New[int, int]()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Set(i, i)
		c.Get(i)
	}
}

func BenchmarkConcurrentReadWrite(b *testing.B) {
	c := New[int, int]()
	for i := 0; i < 1000; i++ {
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

func BenchmarkConcurrentRead(b *testing.B) {
	c := New[int, int]()
	for i := 0; i < 1000; i++ {
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

func BenchmarkSetIfAbsent(b *testing.B) {
	c := New[int, int]()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.SetIfAbsent(i, i)
	}
}

func BenchmarkGetOrSet(b *testing.B) {
	c := New[int, int]()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.GetOrSet(i%100, i)
	}
}

// 大量データテスト

func TestLargeScale(t *testing.T) {
	t.Parallel()

	const n = 100_000
	c := New[int, int]()

	// 大量のSetとGet
	for i := 0; i < n; i++ {
		c.Set(i, i*10)
	}
	if c.Len() != n {
		t.Errorf("Len should return %d after %d Sets, got %d", n, n, c.Len())
	}

	for i := 0; i < n; i++ {
		val, ok := c.Get(i)
		if !ok {
			t.Fatalf("Key %d should exist", i)
		}
		if val != i*10 {
			t.Fatalf("Key %d should have value %d, got %d", i, i*10, val)
		}
	}

	// Keys, Values, Snapshot の件数が正しいこと
	keys := c.Keys()
	if len(keys) != n {
		t.Errorf("Keys should return %d keys, got %d", n, len(keys))
	}

	values := c.Values()
	if len(values) != n {
		t.Errorf("Values should return %d values, got %d", n, len(values))
	}

	items := c.Snapshot()
	if len(items) != n {
		t.Errorf("Snapshot should return %d items, got %d", n, len(items))
	}

	// Drain で全件取り出し
	drained := c.Drain()
	if len(drained) != n {
		t.Errorf("Drain should return %d items, got %d", n, len(drained))
	}
	if c.Len() != 0 {
		t.Errorf("Len should be 0 after Drain, got %d", c.Len())
	}
}

func TestLargeScaleConcurrency(t *testing.T) {
	t.Parallel()

	const n = 100_000
	c := New[int, int]()
	var wg sync.WaitGroup

	// 大量の並行書き込み
	start := make(chan struct{})
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			<-start
			c.Set(id, id*10)
		}(i)
	}
	close(start)
	wg.Wait()

	if c.Len() != n {
		t.Errorf("Len should return %d, got %d", n, c.Len())
	}

	// 大量の並行読み書き混在
	start2 := make(chan struct{})
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			<-start2
			switch id % 4 {
			case 0:
				c.Set(id, id*100)
			case 1:
				c.Get(id)
			case 2:
				c.Has(id)
			case 3:
				c.Delete(id)
			}
		}(i)
	}
	close(start2)
	wg.Wait()
}

func TestLargeScaleOverwrite(t *testing.T) {
	t.Parallel()

	const n = 100_000
	c := New[string, int]()

	// 同一キーに大量上書き
	for i := 0; i < n; i++ {
		c.Set("key", i)
	}
	if c.Len() != 1 {
		t.Errorf("Len should be 1 after overwriting same key %d times, got %d", n, c.Len())
	}
	val, ok := c.Get("key")
	if !ok || val != n-1 {
		t.Errorf("Get should return last written value %d, got %d", n-1, val)
	}

	// 多数のキーをSetしてからDeleteし、件数が一致することを確認
	c2 := New[int, int]()
	for i := 0; i < n; i++ {
		c2.Set(i, i)
	}
	for i := 0; i < n; i += 2 {
		c2.Delete(i)
	}
	expected := n / 2
	if c2.Len() != expected {
		t.Errorf("Len should be %d after deleting half, got %d", expected, c2.Len())
	}

	// 残りのキーが全て正しい値を持つことを確認
	for i := 1; i < n; i += 2 {
		val, ok := c2.Get(i)
		if !ok || val != i {
			t.Fatalf("Key %d should exist with value %d, got %v, %v", i, i, val, ok)
		}
	}
}

func BenchmarkKeys(b *testing.B) {
	for _, size := range []int{100, 1_000, 10_000} {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			c := New[int, int]()
			for i := 0; i < size; i++ {
				c.Set(i, i)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				c.Keys()
			}
		})
	}
}

func BenchmarkSnapshot(b *testing.B) {
	for _, size := range []int{100, 1_000, 10_000} {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			c := New[int, int]()
			for i := 0; i < size; i++ {
				c.Set(i, i)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				c.Snapshot()
			}
		})
	}
}

func TestDeleteFunc(t *testing.T) {
	t.Parallel()

	c := New[string, int]()
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

	// 空のキャッシュに対しても安全に呼べる
	c2 := New[string, int]()
	c2.DeleteFunc(func(k string, v int) bool { return true })
	if c2.Len() != 0 {
		t.Errorf("DeleteFunc on empty cache should be safe, got len %d", c2.Len())
	}

	// 全件削除
	c3 := New[string, int]()
	c3.Set("x", 1)
	c3.Set("y", 2)
	c3.DeleteFunc(func(k string, v int) bool { return true })
	if c3.Len() != 0 {
		t.Errorf("DeleteFunc with always-true should delete all, got len %d", c3.Len())
	}
}

func TestDeleteFuncConcurrency(t *testing.T) {
	t.Parallel()

	c := New[int, int]()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			c.Set(id, id)
		}(i)
	}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.DeleteFunc(func(k, v int) bool { return k%2 == 0 })
		}()
	}

	wg.Wait()
}

func TestPop(t *testing.T) {
	t.Parallel()

	c := New[string, int]()
	c.Set("key1", 42)
	c.Set("key2", 100)

	val, ok := c.Pop("key1")
	if !ok || val != 42 {
		t.Errorf("Pop should return (42, true), got (%v, %v)", val, ok)
	}
	if c.Has("key1") {
		t.Error("Pop should remove the key from cache")
	}
	if c.Len() != 1 {
		t.Errorf("Len should be 1 after Pop, got %d", c.Len())
	}

	// 存在しないキー
	val, ok = c.Pop("not-exists")
	if ok {
		t.Error("Pop should return false for non-existent key")
	}
	if val != 0 {
		t.Errorf("Pop should return zero value for non-existent key, got %v", val)
	}

	// key2はまだ存在する
	val, ok = c.Pop("key2")
	if !ok || val != 100 {
		t.Errorf("Pop should return (100, true), got (%v, %v)", val, ok)
	}
	if c.Len() != 0 {
		t.Errorf("Len should be 0 after popping all keys, got %d", c.Len())
	}
}

func TestPopConcurrency(t *testing.T) {
	t.Parallel()

	c := New[int, int]()
	for i := 0; i < 100; i++ {
		c.Set(i, i*10)
	}

	var wg sync.WaitGroup
	results := make([]bool, 100)
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_, results[id] = c.Pop(id)
		}(i)
	}
	wg.Wait()

	for i := 0; i < 100; i++ {
		if !results[i] {
			t.Errorf("Pop(%d) should have succeeded", i)
		}
	}
	if c.Len() != 0 {
		t.Errorf("Cache should be empty after popping all keys, got len %d", c.Len())
	}
}

func TestSetAll(t *testing.T) {
	t.Parallel()

	c := New[string, int]()
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

	// 上書きも可能
	c.SetAll(map[string]int{"existing": 0})
	val, _ := c.Get("existing")
	if val != 0 {
		t.Errorf("SetAll should overwrite existing keys, got %v", val)
	}

	// 空マップでも安全
	c.SetAll(map[string]int{})
	if c.Len() != 4 {
		t.Errorf("SetAll with empty map should not change cache, got len %d", c.Len())
	}
}

func TestSetAllConcurrency(t *testing.T) {
	t.Parallel()

	c := New[int, int]()
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			c.SetAll(map[int]int{id: id * 10, id + 1000: id * 100})
		}(i)
	}

	wg.Wait()
}

func TestGetAll(t *testing.T) {
	t.Parallel()

	c := New[string, int]()
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

	// キーなしで呼んでも安全
	empty := c.GetAll()
	if len(empty) != 0 {
		t.Errorf("GetAll with no keys should return empty map, got %v", empty)
	}

	// 全て存在しないキー
	none := c.GetAll("x", "y")
	if len(none) != 0 {
		t.Errorf("GetAll with all non-existent keys should return empty map, got %v", none)
	}
}

func TestGetAllConcurrency(t *testing.T) {
	t.Parallel()

	c := New[int, int]()
	for i := 0; i < 100; i++ {
		c.Set(i, i*10)
	}

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_ = c.GetAll(id, id+1, id+2)
		}(i)
	}
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			c.Set(id, id*100)
		}(i)
	}

	wg.Wait()
}

func TestGetOrSetFunc(t *testing.T) {
	t.Parallel()

	c := New[string, int]()

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
	stored, ok := c.Get("key1")
	if !ok || stored != 42 {
		t.Errorf("GetOrSetFunc should persist value, got %v, %v", stored, ok)
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

func TestGetOrSetFuncConcurrency(t *testing.T) {
	t.Parallel()

	c := New[string, int]()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_ = c.GetOrSetFunc("key", func() int { return id })
		}(i)
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

func TestCount(t *testing.T) {
	t.Parallel()

	c := New[string, int]()
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

	n = c.Count(func(k string, v int) bool { return false })
	if n != 0 {
		t.Errorf("Count with always-false should be 0, got %d", n)
	}
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	c := New[string, int]()
	c.Set("counter", 10)

	// 存在するキーを更新
	ok := c.Update("counter", func(v int) int { return v + 1 })
	if !ok {
		t.Error("Update should return true for existing key")
	}
	val, _ := c.Get("counter")
	if val != 11 {
		t.Errorf("Update should increment value, expected 11, got %d", val)
	}

	// 存在しないキーは更新しない
	ok = c.Update("missing", func(v int) int { return v + 1 })
	if ok {
		t.Error("Update should return false for non-existent key")
	}
	if c.Has("missing") {
		t.Error("Update should not create a new key")
	}

	// 複数回更新
	for i := 0; i < 5; i++ {
		c.Update("counter", func(v int) int { return v * 2 })
	}
	val, _ = c.Get("counter")
	if val != 11*32 {
		t.Errorf("Update after 5 doublings, expected %d, got %d", 11*32, val)
	}
}

func TestUpdateConcurrency(t *testing.T) {
	t.Parallel()

	c := New[string, int]()
	c.Set("counter", 0)
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
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

func TestSwap(t *testing.T) {
	t.Parallel()

	c := New[string, int]()
	c.Set("key1", 42)

	// 存在するキーをスワップ
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

	// 存在しないキーをスワップ（新規作成される）
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
	if c.Len() != 2 {
		t.Errorf("Len should be 2, got %d", c.Len())
	}
}

func TestSwapConcurrency(t *testing.T) {
	t.Parallel()

	c := New[string, int]()
	c.Set("key", 0)
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			c.Swap("key", id)
		}(i)
	}
	wg.Wait()

	_, ok := c.Get("key")
	if !ok {
		t.Error("key should exist after concurrent swaps")
	}
}

func TestCompareAndSwap(t *testing.T) {
	t.Parallel()

	eq := func(a, b int) bool { return a == b }
	c := New[string, int]()
	c.Set("key1", 42)

	// 値が一致する場合はスワップ成功
	ok := c.CompareAndSwap("key1", 42, 100, eq)
	if !ok {
		t.Error("CompareAndSwap should return true when values match")
	}
	val, _ := c.Get("key1")
	if val != 100 {
		t.Errorf("CompareAndSwap should set new value 100, got %d", val)
	}

	// 値が一致しない場合はスワップ失敗
	ok = c.CompareAndSwap("key1", 42, 200, eq)
	if ok {
		t.Error("CompareAndSwap should return false when values don't match")
	}
	val, _ = c.Get("key1")
	if val != 100 {
		t.Errorf("CompareAndSwap should not change value, expected 100, got %d", val)
	}

	// 存在しないキー
	ok = c.CompareAndSwap("missing", 0, 1, eq)
	if ok {
		t.Error("CompareAndSwap should return false for non-existent key")
	}
}

func TestCompareAndSwapConcurrency(t *testing.T) {
	t.Parallel()

	eq := func(a, b int) bool { return a == b }
	c := New[string, int]()
	c.Set("key", 0)
	var wg sync.WaitGroup

	// 0→1に変更できるのは1つだけ
	results := make([]bool, 100)
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			results[id] = c.CompareAndSwap("key", 0, 1, eq)
		}(i)
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
	val, _ := c.Get("key")
	if val != 1 {
		t.Errorf("Value should be 1 after CAS, got %d", val)
	}
}

func TestReplace(t *testing.T) {
	t.Parallel()

	c := New[string, int]()

	// 存在しないキーはfalse
	ok := c.Replace("missing", 100)
	if ok {
		t.Error("Replace should return false for non-existent key")
	}
	if c.Has("missing") {
		t.Error("Replace should not create a new key")
	}

	// 存在するキーは上書き
	c.Set("key1", 42)
	ok = c.Replace("key1", 100)
	if !ok {
		t.Error("Replace should return true for existing key")
	}
	val, _ := c.Get("key1")
	if val != 100 {
		t.Errorf("Replace should update value to 100, got %d", val)
	}

	// 削除後はfalse
	c.Delete("key1")
	ok = c.Replace("key1", 200)
	if ok {
		t.Error("Replace should return false after Delete")
	}
}

func TestReplaceConcurrency(t *testing.T) {
	t.Parallel()

	c := New[int, int]()
	for i := 0; i < 100; i++ {
		c.Set(i, i)
	}
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			c.Replace(id, id*100)
		}(i)
	}
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			c.Replace(id+200, 999) // 存在しないキー
		}(i)
	}
	wg.Wait()
}

func TestMap(t *testing.T) {
	t.Parallel()

	c := New[string, int]()
	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)

	// 全ての値を2倍にする
	c.Map(func(k string, v int) int { return v * 2 })

	expected := map[string]int{"a": 2, "b": 4, "c": 6}
	for k, want := range expected {
		got, ok := c.Get(k)
		if !ok || got != want {
			t.Errorf("Map: Get(%q) = (%d, %v), want (%d, true)", k, got, ok, want)
		}
	}

	// アイテム数は変わらない
	if c.Len() != 3 {
		t.Errorf("Map should not change cache size, got %d", c.Len())
	}

	// 空キャッシュでも安全
	c2 := New[string, int]()
	c2.Map(func(k string, v int) int { return v * 2 })
	if c2.Len() != 0 {
		t.Errorf("Map on empty cache should be safe, got len %d", c2.Len())
	}
}

func TestMapConcurrency(t *testing.T) {
	t.Parallel()

	c := New[int, int]()
	for i := 0; i < 100; i++ {
		c.Set(i, i)
	}
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Map(func(k, v int) int { return v + 1 })
		}()
	}
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_, _ = c.Get(id)
		}(i)
	}
	wg.Wait()
}

func TestDeleteAll(t *testing.T) {
	t.Parallel()

	c := New[string, int]()
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

func TestDeleteAllConcurrency(t *testing.T) {
	t.Parallel()

	c := New[int, int]()
	for i := 0; i < 100; i++ {
		c.Set(i, i)
	}
	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			c.DeleteAll(id*5, id*5+1, id*5+2, id*5+3, id*5+4)
		}(i)
	}
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_, _ = c.Get(id)
		}(i)
	}
	wg.Wait()
}

func TestFilter(t *testing.T) {
	t.Parallel()

	c := New[string, int]()
	c.Set("a", 1)
	c.Set("b", -2)
	c.Set("c", 3)
	c.Set("d", -4)

	// 正の値のみ抽出
	result := c.Filter(func(k string, v int) bool { return v > 0 })
	if len(result) != 2 {
		t.Errorf("Filter should return 2 items, got %d", len(result))
	}
	if result["a"] != 1 || result["c"] != 3 {
		t.Errorf("Filter should return positive values, got %v", result)
	}

	// キャッシュは変更されない
	if c.Len() != 4 {
		t.Errorf("Filter should not modify cache, got len %d", c.Len())
	}

	// 返されたマップを変更してもキャッシュに影響しない
	result["a"] = 999
	val, _ := c.Get("a")
	if val != 1 {
		t.Errorf("Modifying Filter result should not affect cache, got %d", val)
	}

	// 何もマッチしない場合
	empty := c.Filter(func(k string, v int) bool { return false })
	if len(empty) != 0 {
		t.Errorf("Filter with always-false should return empty map, got %v", empty)
	}

	// 全てマッチする場合
	all := c.Filter(func(k string, v int) bool { return true })
	if len(all) != 4 {
		t.Errorf("Filter with always-true should return all items, got %d", len(all))
	}
}

func TestFilterConcurrency(t *testing.T) {
	t.Parallel()

	c := New[int, int]()
	for i := 0; i < 100; i++ {
		c.Set(i, i)
	}
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = c.Filter(func(k, v int) bool { return v%2 == 0 })
		}()
	}
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			c.Set(id, id*100)
		}(i)
	}
	wg.Wait()
}

func TestCompareAndDelete(t *testing.T) {
	t.Parallel()

	c := New[string, int]()
	c.Set("key1", 42)

	// 条件を満たす場合は削除
	ok := c.CompareAndDelete("key1", func(v int) bool { return v == 42 })
	if !ok {
		t.Error("CompareAndDelete should return true when condition matches")
	}
	if c.Has("key1") {
		t.Error("CompareAndDelete should remove the key")
	}

	// 条件を満たさない場合は削除しない
	c.Set("key2", 100)
	ok = c.CompareAndDelete("key2", func(v int) bool { return v == 42 })
	if ok {
		t.Error("CompareAndDelete should return false when condition doesn't match")
	}
	if !c.Has("key2") {
		t.Error("CompareAndDelete should not remove key when condition doesn't match")
	}

	// 存在しないキー
	ok = c.CompareAndDelete("missing", func(v int) bool { return true })
	if ok {
		t.Error("CompareAndDelete should return false for non-existent key")
	}
}

func TestCompareAndDeleteConcurrency(t *testing.T) {
	t.Parallel()

	c := New[int, int]()
	for i := 0; i < 100; i++ {
		c.Set(i, i)
	}
	var wg sync.WaitGroup

	// 偶数値のみ削除を並行実行
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			c.CompareAndDelete(id, func(v int) bool { return v%2 == 0 })
		}(i)
	}
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_, _ = c.Get(id)
		}(i)
	}
	wg.Wait()
}

func BenchmarkDrain(b *testing.B) {
	for _, size := range []int{100, 1_000, 10_000} {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			c := New[int, int]()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for j := 0; j < size; j++ {
					c.Set(j, j)
				}
				c.Drain()
			}
		})
	}
}
