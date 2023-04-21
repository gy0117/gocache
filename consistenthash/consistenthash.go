package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// 一致性哈希的实现
// 1. add节点、get节点

// hash算法，外部提供
type HashFunc func(data []byte) uint32

type Map struct {
	hash    HashFunc       // hash算法，外部提供
	replics int            // 每个真实节点的虚拟节点数
	keyring []int          // 哈希环
	hashMap map[int]string // 虚拟节点与真实节点的对应关系
}

func New(replics int, hash HashFunc) *Map {
	m := &Map{
		hash:    hash,
		replics: replics,
		hashMap: make(map[int]string),
	}

	if m.hash == nil {
		// 默认hash算法
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// 添加真实的节点
func (m *Map) Add(keys ...string) {
	for _, v := range keys {
		// v是一个真实的节点
		// 将真实的节点虚拟化
		for i := 0; i < m.replics; i++ {
			// 虚拟节点的hash
			hash := int(m.hash([]byte(strconv.Itoa(i) + v)))
			m.keyring = append(m.keyring, hash)
			m.hashMap[hash] = v
		}
	}
	// 将所有的节点排序
	sort.Ints(m.keyring)
}

// func (m *Map) Get(key string) string {}

// 计算key的哈希值，找到分配的节点
func (m *Map) Get(key string) string {
	if len(m.keyring) == 0 {
		return ""
	}
	hash := int(m.hash([]byte(key)))
	// 二分法找到第一个比hash大的元素，在keyring中
	idx := sort.Search(len(m.keyring), func(i int) bool {
		return m.keyring[i] >= hash
	})
	// 选中节点的hash
	// sort.Search方法：如果没有这样的index，就会返回n，因此这里需要取余
	res := m.keyring[idx%len(m.keyring)]
	return m.hashMap[res]
}
