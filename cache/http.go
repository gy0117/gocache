package cache

import (
	"log"
	"net/http"
	"strings"
)

const CACHE_BASE_PATH = "/_marscache/"

// 分布式缓存，实现节点间通信
type HttpPool struct {
	hostPort string
	basepath string
}

func NewHttpPool(hostport string) *HttpPool {
	return &HttpPool{
		hostPort: hostport,
		basepath: CACHE_BASE_PATH,
	}
}

// 1. 解析url，拿到groupname和key
//		判断path是否是以/_geecache/为前缀的。 否，则panic
// 2. 根据group和key，获取到对应的value，然后写到writer中
// 例如：http://127.0.0.1/_marscache/users/zhangsan，groupname是users，key是zhangsan，即获取users group下的key为zhangsan对应的value
func (p *HttpPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basepath) {
		panic("url path is not '/_geecache/' as prefix.")
	}

	path := r.URL.Path
	// /_marscache/scores/Tom
	log.Printf("path:%v\n", path[len(CACHE_BASE_PATH):])
	parts := strings.SplitN(path[len(CACHE_BASE_PATH):], "/", 2)

	groupname := parts[0]
	key := parts[1]

	log.Printf("group_name: %v, key: %v\n", groupname, key)

	g := GetGroup(groupname)

	item, err := g.Get(key)
	if err != nil {
		log.Printf("g.Get | key: %v, err: %+v\n", key, err)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	// 返回副本
	w.Write(item.ByteSlice())

}
