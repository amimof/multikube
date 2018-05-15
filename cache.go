package multikube

import (
	"github.com/google/uuid"
	"time"
)

// Root cache object
type Cache struct {
	ID uuid.UUID
	client *Request
	Store map[string]Item
}

type Item struct {
	Key string
	Value interface{}
	Created time.Time
}

func (c *Cache) ListKeys() []string {
	keys := make([]string, 0)
	for key, _ := range c.Store {
		keys = append(keys, key)
	}
	return keys
}

func (c *Cache) Get(key string) *Item {
	var item Item
	if c.Exists(key) {
		item = c.Store[key]
	}
	return &item
}

func (c *Cache) Set(key string, val interface{}) *Item {
	item := c.Store[key]
	item.Key = key
	item.Value = val
	item.Created = time.Now()
	c.Store[key] = item
	return &item
}

func (c *Cache) Delete(key string) {
	delete(c.Store, key)
}

func (c *Cache) Exists(key string) bool {
	if _, ok := c.Store[key]; ok {
		return true
	}
	return false
}