# go-cache

A simple, thread-safe, and generic in-memory cache implementation for Go.

## Features

- Thread-safe operations
- Support for any comparable key types and any value types
- Two implementations: `Cache` (sync.Mutex) and `HeavyCache` (sync.RWMutex)
- Simple and lightweight API
- Zero dependencies
- Fully tested with race detector

## Installation

Requires Go 1.22 or later.

```bash
go get github.com/f0reth/go-cache
```

## Cache vs HeavyCache

|                  | `Cache`               | `HeavyCache`              |
| ---------------- | --------------------- | ------------------------- |
| Lock             | `sync.Mutex`          | `sync.RWMutex`            |
| Best for         | Write-heavy workloads | Read-heavy workloads      |
| Read concurrency | Exclusive             | Shared (multiple readers) |

Use `HeavyCache` when reads significantly outnumber writes for better throughput.

## Usage

### Basic Usage

```go
package main

import (
	"fmt"
	"github.com/f0reth/go-cache"
)

func main() {
	// Create a new cache with string keys and int values
	c := cache.New[string, int]()

	// Store values
	c.Set("one", 1)
	c.Set("two", 2)

	// Retrieve values
	val, exists := c.Get("one")
	if exists {
		fmt.Printf("Value: %d\n", val) // Output: Value: 1
	}

	// Delete a value
	c.Delete("one")

	// Check if a key exists
	if c.Has("two") {
		fmt.Println("Key 'two' exists")
	}

	// Get the number of items
	fmt.Printf("Cache size: %d\n", c.Len()) // Output: Cache size: 1

	// Clear all cache entries
	c.Clear()
}
```

### HeavyCache (RWMutex)

```go
// Create a HeavyCache for read-heavy workloads
c := cache.NewHeavy[string, int]()

// Same API as Cache
c.Set("key", 42)
val, ok := c.Get("key") // Uses RLock for concurrent reads
```

### SetIfAbsent / GetOrSet

```go
c := cache.New[string, int]()

// Set only if the key does not exist
ok := c.SetIfAbsent("key", 42) // true (set)
ok = c.SetIfAbsent("key", 99)  // false (already exists, value remains 42)

// Get existing value, or set and return the default
val := c.GetOrSet("key", 99)     // 42 (already exists)
val = c.GetOrSet("new-key", 100) // 100 (set and returned)
```

### GetOrSetFunc

```go
c := cache.New[string, *Config]()

// Like GetOrSet, but lazily initializes the value only when needed
config := c.GetOrSetFunc("app", func() *Config {
	return loadConfigFromDisk() // Only called if key doesn't exist
})
```

### Batch Operations

```go
c := cache.New[string, int]()

// Set multiple items at once
c.SetAll(map[string]int{"a": 1, "b": 2, "c": 3})

// Get multiple items at once (only existing keys are returned)
items := c.GetAll("a", "b", "missing") // map[string]int{"a": 1, "b": 2}

// Delete multiple keys at once
c.DeleteAll("a", "b")
```

### Keys / Values / Snapshot

```go
c := cache.New[string, int]()
c.Set("a", 1)
c.Set("b", 2)

keys := c.Keys()       // []string{"a", "b"} (order not guaranteed)
values := c.Values()   // []int{1, 2} (order not guaranteed)
items := c.Snapshot()   // map[string]int{"a": 1, "b": 2} (copy)
```

### Drain

```go
c := cache.New[string, int]()
c.Set("a", 1)
c.Set("b", 2)

// Drain returns all items and empties the cache
items := c.Drain() // map[string]int{"a": 1, "b": 2}
fmt.Println(c.Len()) // Output: 0
```

### Pop / Swap / Replace

```go
c := cache.New[string, int]()
c.Set("key", 42)

// Pop retrieves and removes a value
val, ok := c.Pop("key") // 42, true
val, ok = c.Pop("key")  // 0, false

// Swap replaces a value and returns the old one
c.Set("key", 1)
old, ok := c.Swap("key", 2) // 1, true

// Replace overwrites only if the key already exists
ok = c.Replace("key", 3)       // true
ok = c.Replace("missing", 3)   // false
```

### Update / CompareAndSwap / CompareAndDelete

```go
c := cache.New[string, int]()
c.Set("counter", 10)

// Update a value in-place
c.Update("counter", func(v int) int { return v + 1 }) // 11

// CompareAndSwap replaces only if the current value matches
c.CompareAndSwap("counter", 11, 20, func(a, b int) bool { return a == b }) // true

// CompareAndDelete removes only if the value satisfies the condition
c.CompareAndDelete("counter", func(v int) bool { return v > 10 }) // true
```

### Iteration and Filtering

```go
c := cache.New[string, int]()
c.SetAll(map[string]int{"a": 1, "b": 2, "c": 3})

// Range iterates over all items (return false to stop early)
c.Range(func(key string, value int) bool {
	fmt.Printf("%s: %d\n", key, value)
	return true
})

// Count items matching a condition
n := c.Count(func(k string, v int) bool { return v > 1 }) // 2

// Filter returns a snapshot of matching items
evens := c.Filter(func(k string, v int) bool { return v%2 == 0 })

// DeleteFunc removes all items matching a condition
c.DeleteFunc(func(k string, v int) bool { return v < 2 })

// Map transforms all values in-place
c.Map(func(k string, v int) int { return v * 10 })
```

### Custom Key and Value Types

```go
package main

import (
	"fmt"
	"github.com/f0reth/go-cache"
)

type User struct {
	ID   int
	Name string
	Age  int
}

func main() {
	// Cache with int keys and User values
	userCache := cache.New[int, User]()

	// Store user objects
	userCache.Set(1, User{ID: 1, Name: "Alice", Age: 30})
	userCache.Set(2, User{ID: 2, Name: "Bob", Age: 25})

	// Retrieve a user
	user, exists := userCache.Get(1)
	if exists {
		fmt.Printf("User: %+v\n", user) // Output: User: {ID:1 Name:Alice Age:30}
	}
}
```

## API

Both `Cache` and `HeavyCache` implement `CacheInterface[K, V]`:

| Method                                   | Description                              |
| ---------------------------------------- | ---------------------------------------- |
| `Set(key, value)`                        | Store a value                            |
| `Get(key) (V, bool)`                     | Retrieve a value                         |
| `Delete(key)`                            | Delete a value                           |
| `Clear()`                                | Remove all entries                       |
| `Len() int`                              | Return the number of items               |
| `Has(key) bool`                          | Check if a key exists                    |
| `Keys() []K`                             | Return all keys                          |
| `Values() []V`                           | Return all values                        |
| `Snapshot() map[K]V`                     | Return a copy of all items               |
| `Drain() map[K]V`                        | Return all items and empty the cache     |
| `SetIfAbsent(key, value) bool`           | Set only if the key does not exist       |
| `GetOrSet(key, value) V`                 | Get existing or set and return default   |
| `GetOrSetFunc(key, fn) V`                | Like GetOrSet but lazily computes value  |
| `SetAll(items)`                          | Set multiple items at once               |
| `GetAll(keys...) map[K]V`                | Get multiple items at once               |
| `DeleteAll(keys...)`                     | Delete multiple keys at once             |
| `Pop(key) (V, bool)`                     | Get and remove a value                   |
| `Swap(key, value) (V, bool)`             | Replace value and return old one         |
| `Replace(key, value) bool`               | Overwrite only if key exists             |
| `Update(key, fn) bool`                   | Update value in-place with a function    |
| `CompareAndSwap(key, old, new, eq) bool` | Replace only if current value matches    |
| `CompareAndDelete(key, fn) bool`         | Delete only if value satisfies condition |
| `Range(fn)`                              | Iterate over all items                   |
| `Count(fn) int`                          | Count items matching a condition         |
| `Filter(fn) map[K]V`                     | Return snapshot of matching items        |
| `DeleteFunc(fn)`                         | Delete items matching a condition        |
| `Map(fn)`                                | Transform all values in-place            |

## License

MIT
