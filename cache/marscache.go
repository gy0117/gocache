package cache

import (
	"fmt"
	"log"
	"sync"

	"github.com/marsxingzhi/marscache/pb"
	"github.com/marsxingzhi/marscache/peers"
	"github.com/marsxingzhi/marscache/singleflight"
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
	name       string
	getter     Getter
	mainCache  cacheInner
	peerPicker peers.PeerPicker

	loader *singleflight.Group
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
		loader:    &singleflight.Group{},
	}

	mutex.Lock()
	groups[name] = g
	mutex.Unlock()
	return g
}

func (g *Group) RegisterPeerPicker(peer peers.PeerPicker) {
	if g.peerPicker != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peerPicker = peer
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
		log.Printf("Group.Get | mainCache.get successfully data: %v\n", bytedata.String())
		return bytedata, nil
	}
	log.Println("Group.Get | cache miss")
	// 如果没有缓存，则加载本地或者远程的
	// return g.load(key)

	data, err := g.loader.Do(key, func() (singleflight.CallValue, error) {
		return g.load(key)
	})
	if err != nil {
		return ByteData{}, err
	}
	return data.(ByteData), nil
}

func (g *Group) put(key string, value ByteData) {
	g.mainCache.add(key, value)
}

// 1. 先去远程查找
// 2. 远程找不到，再去本地找
func (g *Group) load(key string) (ByteData, error) {
	if g.peerPicker != nil {
		if peer, ok := g.peerPicker.PickPeer(key); ok {
			bytedata, err := g.GetFromPeerPicker(peer, key)
			if err == nil {
				log.Printf("Group.load | get from PeerPicker successfully, data: %+v\n", bytedata.String())
				return bytedata, nil
			}
			log.Printf("Group.load | failed to get from peer, failed: %+v\n", err)
		}
	}
	return g.loadLocally(key)
}

func (g *Group) loadLocally(key string) (ByteData, error) {
	log.Printf("Group.loadLocally | key: %v\n", key)
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

func (g *Group) GetFromPeerPicker(peerGetter peers.PeerGetter, key string) (ByteData, error) {
	req := &pb.Request{
		Group: g.name,
		Key:   key,
	}

	resp := &pb.Response{}

	// b, err := peerGetter.Get(g.name, key)

	err := peerGetter.Get(req, resp)

	if err != nil {
		return ByteData{}, err
	}
	return ByteData{data: resp.Value}, nil
}
