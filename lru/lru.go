package lru

import (
	"container/list"
	"sync"
)

type Cache struct {
	mu sync.Mutex

	//k -> element(key, value)
	mem map[string]*list.Element

	//capacity of cache
	capacity int

	//doubly LL for LRU Policy -> holds keys so we know which kvp to remove
	lru_list *list.List

	//cache should have a unique id on creation (cache ip hash) -> maybe?
	cache_ip int
}

// this what the doubly ll stores in each node asw as the val in the map
type entry struct {
	key   string
	value string
}

func NewCache(capacity int) *Cache {
	return &Cache{
		mem:      make(map[string]*list.Element),
		capacity: capacity,
		lru_list: list.New(),
	}

}

// use for puts -> should handle case where cache cap is full for api simplicity
func CachePut(cache *Cache, newkey string, newval string) {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	if len(cache.mem) == cache.capacity {
		lastElement := cache.lru_list.Back()
		e := lastElement.Value.(entry)
		key := e.key
		delete(cache.mem, key)
		cache.lru_list.Remove(lastElement)
	}
	elem := cache.lru_list.PushFront(entry{newkey, newval})
	cache.mem[newkey] = elem
}

// use for gets (reads so lru policy moves this elem to front)
func CacheGet(cache *Cache, key string) string {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	if e, ok := cache.mem[key]; ok {
		//getting element->casting into entry struct -> getting val -> moving to front -> return val
		elem := e.Value.(entry)
		val := elem.value
		cache.lru_list.MoveToFront(e)
		return val
		//TODO -> instead of returning key not found, look in db and then update cache to have that key
	} else {
		return "key not found"
	}
}
