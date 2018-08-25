package multikube

import (
	"github.com/google/uuid"
	"time"
)

// Root cache object
type Cache struct {
	ID    uuid.UUID
	Store map[string]Item
	TTL		time.Duration
}

// Item represents a unit stored in the cache
type Item struct {
	Key     string
	Value   []byte
	Created time.Time
	Updated time.Time
}

// ListKeys returns the keys of all items in the cache as a string array
func (c *Cache) ListKeys() []string {
	keys := make([]string, 0)
	for key := range c.Store {
		keys = append(keys, key)
	}
	return keys
}

// Get returns an item from the cache by key
func (c *Cache) Get(key string) *Item {
	var item Item
	if c.Exists(key) {
		item = c.Store[key]
	}
	return &item
}

// Set instantiates and allocates a key in the cache and overwrites any previously set item
func (c *Cache) Set(key string, val []byte) *Item {
	item := c.Store[key]
	item.Key = key
	item.Value = val
	// TODO: Only set Created timestamp once, not for every update
	item.Created = time.Now()
	item.Updated = time.Now()
	c.Store[key] = item
	return &item
}

// Delete removes an item by key
func (c *Cache) Delete(key string) {
	delete(c.Store, key)
}

// Exists returns true if an item with the given exists is non-nil. Otherwise returns false
func (c *Cache) Exists(key string) bool {
	if _, ok := c.Store[key]; ok {
		return true
	}
	return false
}

// Len returns the number of items stored in cache
func (c *Cache) Len() int {
	return len(c.Store)
}

// Size return the sum of all bytes in the cache
func (c *Cache) Size() int {
	l := 0
	for _, val := range c.Store {
		l += val.Bytes()
	}
	return l
}

// Age returns the duration elapsed since creation
func (i *Item) Age() time.Duration {
	return time.Now().Sub(i.Created)
}


// Byte returns the number of bytes of i. Shorthand for len(i.Value)
func (i *Item) Bytes() int {
	return len(i.Value)
}

// NewCache return a new empty cache
func NewCache() *Cache {
	return &Cache{
		ID:    uuid.New(),
		Store: make(map[string]Item),
		TTL:	 time.Second*1,
	}
}
