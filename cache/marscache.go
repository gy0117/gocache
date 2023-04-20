package cache

import (
	"fmt"
	"sync"
)

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

// 接口型函数
func (gf GetterFunc) Get(key string) ([]byte, error) {
	return gf(key)
}

type Group struct {
	name      string
	getter    Getter
	mainCache cacheInner
}

var (
	mutex  sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, capacity int64, getter Getter) *Group {
	if getter == nil {
		panic("Getter is nil")
	}

	cache := cacheInner{
		cacheCapacity: capacity,
	}
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache,
	}

	mutex.Lock()
	groups[name] = g
	mutex.Unlock()
	return g
}

func GetGroup(name string) *Group {
	mutex.RLock()
	g := groups[name]
	mutex.RUnlock()
	return g
}

func (g *Group) Get(key string) (ByteData, error) {
	if key == "" {
		return ByteData{}, fmt.Errorf("key must not be nil")
	}

	if bytedata, ok := g.mainCache.get(key); ok {
		return bytedata, nil
	}
	// 如果没有缓存，则加载本地或者远程的
	return g.load(key)
}

func (g *Group) put(key string, value ByteData) {
	g.mainCache.add(key, value)
}

func (g *Group) load(key string) (ByteData, error) {
	return g.loadLocally(key)
}

func (g *Group) loadLocally(key string) (ByteData, error) {
	bytedata, err := g.getter.Get(key)
	if err != nil {
		return ByteData{}, err
	}

	val := ByteData{
		data: cloneBytes(bytedata),
	}

	// 添加到缓存
	g.put(key, val)
	return val, nil
}
