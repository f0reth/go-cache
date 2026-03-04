package cache

import (
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

	// 複数のゴルーチンで同時に書き込み
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			c.Set(id, id*10)
		}(i)
	}

	// 複数のゴルーチンで同時に読み込み
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_, _ = c.Get(id)
		}(i)
	}

	// 一部のキーを削除
	for i := 0; i < 50; i += 2 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			c.Delete(id)
		}(i)
	}

	// すべてのゴルーチンの完了を待つ
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
}
