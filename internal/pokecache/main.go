package pokecache

import(
	"time"
	"sync"
	//"fmt"
)

var (
    Mu   sync.Mutex
)

type Pokecache struct {
	Entries map[string]cacheEntry
	Interval time.Duration
}

type cacheEntry struct {
	createdAt time.Time
	val []byte
}

func NewCache(interval time.Duration) Pokecache {
	new_cache := Pokecache{Interval: interval,Entries: make(map[string]cacheEntry)}
	go new_cache.reapLoop()
	return new_cache
}

func (cache *Pokecache) Add(key string, val []byte){
	Mu.Lock()
	defer Mu.Unlock()
	cache.Entries[key] = cacheEntry{val: val,createdAt: time.Now()}
}

func (cache *Pokecache) Get(key string) ([]byte,bool) {
	Mu.Lock()
	defer Mu.Unlock()
	entry,exists := cache.Entries[key]
	
	return entry.val,exists
}

func (cache *Pokecache) reapLoop() { //interval time.Duration//
	
	
	interval := cache.Interval
	ticker := time.NewTicker(interval)
	for range ticker.C {
		Mu.Lock()
		for k,entry := range cache.Entries {			
			if time.Since(entry.createdAt) > interval {
				delete(cache.Entries,k)
			}
		}
		Mu.Unlock()
	}

}