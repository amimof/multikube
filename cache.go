package multikube

import (
	"sync"
	"time"
)

// Root cache object
type Cache struct {
	Store map[string]*Item
	TTL   time.Duration
	mux   sync.Mutex
}

// Item represents a unit stored in the cache
type Item struct {
	Key     string
	Value   []byte
	expires time.Time
	created time.Time
}

// ListKeys returns the keys of all items in the cache as a string array
func (c *Cache) ListKeys() []string {
	c.mux.Lock()
	defer c.mux.Unlock()

	keys := make([]string, 0, len(c.Store))
	for key := range c.Store {
		keys = append(keys, key)
	}
	return keys
}

// Get returns an item from the cache by key
func (c *Cache) Get(key string) *Item {
	c.mux.Lock()
	defer c.mux.Unlock()

	if !c.Exists(key) {
		return nil
	}

	item := c.Store[key]
	if !item.expires.IsZero() && item.Age() > c.TTL {
		// Item age exceeded time to live
		delete(c.Store, key)
		return nil
	}
	return item
}

// Set instantiates and allocates a key in the cache and overwrites any previously set item
func (c *Cache) Set(key string, val []byte) *Item {
	c.mux.Lock()
	defer c.mux.Unlock()

	// Delete item if already exists in the cache
	if c.Exists(key) {
		delete(c.Store, key)
	}

	item := &Item{
		Key:     key,
		Value:   val,
		expires: time.Now().Add(c.TTL),
		created: time.Now(),
	}

	c.Store[key] = item
	return item
}

// Delete removes an item by key
func (c *Cache) Delete(key string) {
	c.mux.Lock()
	defer c.mux.Unlock()
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
	c.mux.Lock()
	defer c.mux.Unlock()

	return len(c.Store)
}

// Size return the sum of all bytes in the cache
func (c *Cache) Size() int {
	c.mux.Lock()
	defer c.mux.Unlock()

	l := 0
	for _, val := range c.Store {
		l += val.Bytes()
	}
	return l
}

// Age returns the duration elapsed since creation
func (i *Item) Age() time.Duration {
	return time.Now().Sub(i.created)
}

// ExpiresAt return the time when the item was created plus the configured TTL
func (i *Item) ExpiresAt() time.Time {
	return i.expires
}

// Byte returns the number of bytes of i. Shorthand for len(i.Value)
func (i *Item) Bytes() int {
	return len(i.Value)
}

// NewCache return a new empty cache
func NewCache() *Cache {
	return &Cache{
		Store: make(map[string]*Item),
		TTL:   time.Second * 1,
	}
}
