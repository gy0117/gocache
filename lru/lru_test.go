package lru

import (
	"log"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

const MAX_CAPACITY = 1024 * 1024 * 100

func TestAddAndGet(t *testing.T) {
	convey.Convey("TestAddAndGet", t, func() {
		cache := New(MAX_CAPACITY)
		cache.SetAddHandler(func(s string, v Value) {
			log.Printf("TestAddAndGet | add handler | key: %v, value: %+v\n", s, v)
		})

		convey.Convey("TestAddAndGet success", func() {
			cache.Add("username", String("marsxingzhi"))

			data, ok := cache.Get("username")
			convey.So(ok, convey.ShouldBeTrue)

			want := String("marsxingzhi")
			convey.So(data, convey.ShouldEqual, want)

		})

		convey.Convey("TestAddAndGet failed", func() {
			cache.Add("uid", String("123456"))

			data, ok := cache.Get("uid")
			convey.So(ok, convey.ShouldBeTrue)

			want := String("xingzhi123")
			convey.So(data, convey.ShouldNotEqual, want)

		})
	})
}

type String string

func (str String) Len() int {
	return len(str)
}
