package pokecache

import (
	"sync"
	"time"
)

type cacheEntry struct {
	createdAt time.Time
	val []byte
}

type Cache struct {
	mux *sync.Mutex
	items map[string]cacheEntry
}

func (c Cache) Add(key string, val []byte)  {
	c.mux.Lock()
	c.items[key] = cacheEntry{ time.Now(), val }
	c.mux.Unlock()
}

func (c Cache) reapLoop(interval time.Duration)  {
	ticker := time.NewTicker(interval)
	for {
		t := <-ticker.C
		for k, v := range c.items {
			if t.Sub(v.createdAt) >= interval {
				c.mux.Lock()
				delete(c.items, k)
				c.mux.Unlock()
			}
		}
	}
}

func (c Cache) Get(key string) ([]byte, bool) {
	c.mux.Lock()
	v, ok := c.items[key]
	c.mux.Unlock()

	if ok {
		return v.val, true
	}
	return nil, false
}

func NewCache(interval time.Duration) Cache {
	theMux := sync.Mutex{}
	theCache := Cache{ &theMux, make(map[string]cacheEntry) }
	go theCache.reapLoop(interval)
	return theCache
}
