package cache

import (
	"sync"

	"github.com/marsxingzhi/marscache/lru"
)

// 封装lru，提供并发能力
type cacheInner struct {
	mutex         sync.Mutex
	lru           *lru.Cache
	cacheCapacity int64
}

func (c *cacheInner) add(key string, value ByteData) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.lru == nil {
		c.lru = lru.New(c.cacheCapacity)
	}

	c.lru.Add(key, value)
}

func (c *cacheInner) get(key string) (value ByteData, ok bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.lru == nil {
		return
	}
	if val, ok := c.lru.Get(key); ok {
		return val.(ByteData), ok
	}
	return
}
