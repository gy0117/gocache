package cache

import (
	"fmt"
	"log"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestGetter(t *testing.T) {
	convey.Convey("TestGetter", t, func() {

		convey.Convey("success", func() {

			gf := GetterFunc(func(key string) ([]byte, error) {
				return []byte(key), nil
			})

			want := []byte("key")

			b, _ := gf.Get("key")

			// // a deep equals for arrays, slices, maps, and structs
			convey.So(b, convey.ShouldResemble, want)
		})

	})
}

func TestGet(t *testing.T) {
	convey.Convey("TestGet", t, func() {

		var db = map[string]string{
			"Tom":  "1",
			"Jack": "2",
			"Sam":  "3",
		}

		loadCounts := make(map[string]int, len(db))

		gee := NewGroup("test", 1024*1024*10, GetterFunc(func(key string) ([]byte, error) {
			log.Println("GetterFunc search key ", key)
			if v, ok := db[key]; ok {
				loadCounts[key]++
				return []byte(v), nil
			} else {
				return nil, fmt.Errorf("%s not exist.", key)
			}

		}))

		convey.Convey("test bytedata success", func() {

			for k, v := range db {
				bytedata, _ := gee.Get(k)
				log.Println("GetterFunc search bytedata ", bytedata.String())

				convey.So(bytedata.String(), convey.ShouldEqual, v)
			}
		})

		convey.Convey("test loadLocally", func() {

			for k, _ := range db {
				gee.Get(k)
				convey.So(loadCounts[k] == 1, convey.ShouldBeTrue)
			}
		})
	})
}
