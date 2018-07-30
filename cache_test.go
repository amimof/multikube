package multikube_test

import (
	"gitlab.com/amimof/multikube"
	"testing"
	//"github.com/stretchr/testify/assert"
)

func TestCacheGetItem(t *testing.T) {
	cache := multikube.NewCache()
	item := cache.Get("namespaces")
	t.Logf("Key: %s", item.Key)
	t.Logf("Value: %s", item.Value)
	t.Logf("Created: %s", item.Created.String())
	t.Logf("Updated: %s", item.Updated.String())
}

func TestCacheSetItem(t *testing.T) {
	cache := multikube.NewCache()
	item := cache.Set("namespaces", []byte("Hello World"))
	t.Logf("Key: %s", item.Key)
	t.Logf("Value: %s", item.Value)
	t.Logf("Created: %s", item.Created.String())
	t.Logf("Updated: %s", item.Updated.String())
}

func TestCacheDeleteItem(t *testing.T) {
	cache := multikube.NewCache()

	item := cache.Set("namespaces", []byte("Hello World"))
	t.Logf("Existing:")
	t.Logf("  Key: %s", item.Key)
	t.Logf("  Value: %s", item.Value)
	t.Logf("  Created: %s", item.Created.String())
	t.Logf("  Updated: %s", item.Updated.String())

	cache.Delete(item.Key)
	item = cache.Get("namespaces")
	t.Logf("Deleted:")
	t.Logf("  Key: %s", item.Key)
	t.Logf("  Value: %s", item.Value)
	t.Logf("  Created: %s", item.Created.String())
	t.Logf("  Updated: %s", item.Updated.String())
}

func TestCacheListKeys(t *testing.T) {
	cache := multikube.NewCache()
	cache.Set("/namespaces/", []byte{'a'})
	cache.Set("/namespaces/pods", []byte{'b'})
	cache.Set("/namespaces/pods/pod-1", []byte{'c'})
	for i, key := range cache.ListKeys() {
		t.Logf("[%d]: %s", i, key)
	}
}

func TestCacheSize(t *testing.T) {
	cache := multikube.NewCache()
	cache.Set("A", []byte("foo"))
	t.Logf("Cache size is %d bytes", cache.Size())
}

func TestCacheBytes(t *testing.T) {
	cache := multikube.NewCache()
	cache.Set("A", []byte("foo"))
	cache.Set("B", []byte("bar"))

	a := cache.Get("A")
	b := cache.Get("B")

	t.Logf("Item %s is %d bytes", a.Key, a.Bytes())
	t.Logf("Item %s is %d bytes", b.Key, b.Bytes())
	t.Logf("Items are %d bytes in total", cache.Size())
}

func TestCacheLen(t *testing.T) {
	cache := multikube.NewCache()
	cache.Set("A", []byte("alpha"))
	cache.Set("B", []byte("bravo"))
	cache.Set("C", []byte("charlie"))

	t.Logf("Cache length is %d", cache.Len())
}
