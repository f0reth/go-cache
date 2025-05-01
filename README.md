# go-cache

A simple, thread-safe, and generic in-memory cache implementation for Go.

## Features

- Thread-safe operations
- Support for any comparable key types and any value types
- Simple and lightweight API
- Zero dependencies
- Fully tested

## Installation

Requires Go 1.20 or later.

```bash
go get github.com/daichi2mori/go-cache
```

## Usage

### Basic Usage

```go
package main

import (
	"fmt"
	"github.com/daichi2mori/go-cache"
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

	// Check if value exists after deletion
	_, exists = c.Get("one")
	fmt.Printf("Key 'one' exists: %v\n", exists) // Output: Key 'one' exists: false

	// Clear all cache entries
	c.Clear()
}
```

### Custom Key and Value Types

```go
package main

import (
	"fmt"
	"github.com/daichi2mori/go-cache"
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
