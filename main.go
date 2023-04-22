package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/marsxingzhi/marscache/cache"
)

var db = map[string]string{
	"zhangsan": "100",
	"lisi":     "200",
	"wangwu":   "300",
}

func main() {
	fmt.Println("hello world")

	// 测试
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "marscache server port")
	flag.BoolVar(&api, "api", false, "Start api server?")
	flag.Parse()

	apiAddr := "http://127.0.0.1:9999"
	addrMap := map[int]string{
		8001: "http://127.0.0.1:8001",
		8002: "http://127.0.0.1:8002",
		8003: "http://127.0.0.1:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	group := createGroup()

	if api {
		go startApiServer(apiAddr, group)
	}

	startCacheServer(addrMap[port], []string(addrs), group)

}

// 缓存服务器走的是addr这个请求
// 存在好几个节点addrs，但是这个服务走的是addr
func startCacheServer(addr string, addrs []string, group *cache.Group) {
	peers := cache.NewHttpPool(addr)
	peers.Set(addrs...)
	group.RegisterPeerPicker(peers)
	log.Println("marscache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

func startApiServer(apiAddr string, group *cache.Group) {
	http.Handle("/api", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		log.Printf("startApiServer | query api | key: %v:\n", key)
		bytedata, err := group.Get(key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(bytedata.ByteSlice())
	}))

	log.Println("fontend server is running at ", apiAddr)
	// apiAddr[7:] 去掉 http://
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

func createGroup() *cache.Group {
	return cache.NewGroup("scores", 2<<10, cache.GetterFunc(func(key string) ([]byte, error) {
		log.Printf("DB | query key: %v\n", key)
		if v, ok := db[key]; ok {
			return []byte(v), nil
		}
		return nil, fmt.Errorf("%s not exist", key)
	}))
}

func test() {
	var db = map[string]string{
		"zhangsan": "100",
		"lisi":     "200",
		"wangwu":   "300",
	}
	loadCounts := make(map[string]int, len(db))

	cache.NewGroup("scores", 1024*1024*10, cache.GetterFunc(func(key string) ([]byte, error) {
		log.Println("GetterFunc search key ", key)
		if v, ok := db[key]; ok {
			loadCounts[key]++
			return []byte(v), nil
		} else {
			return nil, fmt.Errorf("%s not exist", key)
		}

	}))

	// for k, _ := range db {
	// 	bytedata, _ := gee.Get(k)
	// 	log.Println("GetterFunc search bytedata ", bytedata.String())
	// }

	addr := "127.0.0.1:8081"
	peers := cache.NewHttpPool(addr)

	err := http.ListenAndServe(addr, peers)
	log.Fatal(err)
}
