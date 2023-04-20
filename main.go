package main

import (
	"fmt"
	"log"

	"github.com/marsxingzhi/marscache/cache"
)

func main() {
	fmt.Println("hello world")

	var db = map[string]string{
		"Tom":  "1",
		"Jack": "2",
		"Sam":  "3",
	}
	loadCounts := make(map[string]int, len(db))

	gee := cache.NewGroup("test", 1024*1024*10, cache.GetterFunc(func(key string) ([]byte, error) {
		log.Println("GetterFunc search key ", key)
		if v, ok := db[key]; ok {
			loadCounts[key]++
			return []byte(v), nil
		} else {
			return nil, fmt.Errorf("%s not exist.", key)
		}

	}))

	for k, _ := range db {
		bytedata, _ := gee.Get(k)
		log.Println("GetterFunc search bytedata ", bytedata.String())
	}

}
