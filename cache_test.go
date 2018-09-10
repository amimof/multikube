package multikube_test

import (
	"gitlab.com/amimof/multikube"
	"testing"
)

var key string = "somekey"
var val string = "hello world"

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func TestCacheGetNilItem(t *testing.T) {
	cache := multikube.NewCache()
	item := cache.Get(key)
	if item != nil {
		t.Fatalf("Item with key %s should be nil", key)
	}
}

func TestCacheSetItem(t *testing.T) {
	cache := multikube.NewCache()
	item := cache.Set(key, []byte(val))
	if item.Key != key {
		t.Fatalf("Item key is not %s", key)
	}
	if string(item.Value) != val {
		t.Fatalf("Item val is not %s", val)
	}
}

func TestCacheSetGetItem(t *testing.T) {
	cache := multikube.NewCache()
	cache.Set(key, []byte(val))
	item := cache.Get(key)
	if item.Key != key {
		t.Fatalf("Item key is not %s", key)
	}
	if string(item.Value) != val {
		t.Fatalf("Item val is not %s", val)
	}
}

func TestCacheDeleteItem(t *testing.T) {
	cache := multikube.NewCache()
	item := cache.Set(key, []byte(val))

	cache.Delete(item.Key)
	item = cache.Get(key)
	if item != nil {
		t.Fatalf("Item with key %s should be nil", key)
	}
}

func TestCacheListKeys(t *testing.T) {
	cache := multikube.NewCache()
	items := []string{"/namespaces/", "/namespaces/pods", "/namespaces/pods/pod-1"}

	cache.Set(items[0], []byte{'a'})
	cache.Set(items[1], []byte{'b'})
	cache.Set(items[2], []byte{'c'})

	for i, k := range cache.ListKeys() {
		if !contains(items, k) {
			t.Fatalf("Key should be %s but got %s", items[i], k)
		}
	}

}

func TestCacheSize(t *testing.T) {
	cache := multikube.NewCache()
	cache.Set(key, []byte("a"))
	if cache.Size() != 1 {
		t.Fatalf("Expected cache size to be %d but got %d", 1, cache.Size())
	}
}

func TestCacheItemBytes(t *testing.T) {
	cache := multikube.NewCache()
	cache.Set("A", []byte("a"))
	cache.Set("B", []byte("b"))
	cache.Set("C", []byte("c"))

	a := cache.Get("A")
	b := cache.Get("B")
	c := cache.Get("C")

	items := []*multikube.Item{a, b, c}

	for _, item := range items {
		if item.Bytes() != 1 {
			t.Fatalf("Expected item bytes to be %d but got %d", 1, item.Bytes())
		}
	}

}

func TestCacheLen(t *testing.T) {
	cache := multikube.NewCache()
	cache.Set("A", []byte("alpha"))
	cache.Set("B", []byte("bravo"))
	cache.Set("C", []byte("charlie"))

	if cache.Len() != 3 {
		t.Logf("Expected cache length to be %d but got %d", 3, cache.Len())
	}

}
