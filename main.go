package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/marsxingzhi/marscache/cache"
)

func main() {
	fmt.Println("hello world")

	var db = map[string]string{
		"Tom":  "10",
		"Jack": "20",
		"Sam":  "30",
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
