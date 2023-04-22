package singleflight

import (
	"sync"
)

// singleflight实现

type CallValue interface{}

// call 表示一个请求
type call struct {
	wg  sync.WaitGroup
	val CallValue
	err error
}

type Group struct {
	mutex sync.Mutex
	calls map[string]*call // 一个key对应一个call
}

type DoFunc func() (CallValue, error)

func (g *Group) Do(key string, doFunc DoFunc) (CallValue, error) {
	g.mutex.Lock()

	if g.calls == nil {
		g.calls = make(map[string]*call)
	}

	if call, ok := g.calls[key]; ok {
		g.mutex.Unlock()

		call.wg.Wait()
		return call.val, call.err
	}

	c := new(call)
	c.wg.Add(1)
	g.calls[key] = c

	g.mutex.Unlock()

	c.val, c.err = doFunc()
	c.wg.Done()

	// 删除call
	// 首先不同的key可能每次的doFunc不一样，因此值是不一样的，所以这里没必要存储；
	// 另外这里也不应该存储数据
	g.mutex.Lock()
	delete(g.calls, key)
	g.mutex.Unlock()

	return c.val, c.err
}
