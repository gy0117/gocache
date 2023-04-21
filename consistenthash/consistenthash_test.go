package consistenthash

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestHash(t *testing.T) {
	convey.Convey("TestHash", t, func() {

		// 假设有三个节点2,4,6，总的节点为[02, 04, 06, 12, 14, 16, 22, 24, 26]
		m := New(3, func(data []byte) uint32 {
			// 模拟的哈希算法
			i, _ := strconv.Atoi(string(data))
			return uint32(i)
		})

		// 先Add
		m.Add("2", "4", "6")

		//
		cases := map[string]string{
			"2":  "2", // key为2，选中的虚拟节点为02，真实节点是2
			"11": "2", // key为11，选中的虚拟节点为12，真实节点是2
			"23": "4", // key为23，选中的虚拟节点为24，真实节点是4
			"27": "2", // key为27，选中的虚拟节点为02， 真实节点是2
			"15": "6", // key为15，选中的虚拟额几点为16，真实节点是6
		}

		for k, v := range cases {
			fmt.Printf("k: %v, v: %v\n", k, v)
			got := m.Get(k)
			convey.So(got, convey.ShouldEqual, v)
		}
	})
}
