package utils

import (
	"bytes"
	"github.com/gy0117/gocache/pkg/utils/codec"
	"math/rand"
	"sync"
)

const (
	defaultMaxLevel = 32
)

type Element struct {
	levels []*Element
	entry  *codec.Entry
	score  float64
}

func newElement(score float64, entry *codec.Entry, level int) *Element {
	return &Element{
		levels: make([]*Element, level+1), // 需要算上原始链表的
		entry:  entry,
		score:  score,
	}
}

type SkipList struct {
	header   *Element
	rand     *rand.Rand
	maxLevel int
	length   int
	size     int64
	sync.RWMutex
}

func NewSkipList() *SkipList {
	header := &Element{
		levels: make([]*Element, defaultMaxLevel),
	}
	list := &SkipList{
		header:   header,
		maxLevel: defaultMaxLevel - 1, // 从0开始
	}
	return list
}

func (list *SkipList) Put(data *codec.Entry) error {
	list.Lock()
	defer list.Unlock()

	prevs := make([]*Element, list.maxLevel+1)
	keyScore := list.calcScore(data.Key)
	prev := list.header

	for i := list.maxLevel; i >= 0; i-- {
		for next := prev.levels[i]; next != nil; next = prev.levels[i] {
			v := list.compare(keyScore, data.Key, next)
			if v <= 0 {
				if v == 0 {
					next.entry = data
					return nil
				}
				break
			}
			prev = next
		}
		prevs[i] = prev
	}

	randLevel := list.randLevel()
	newElem := newElement(keyScore, data, randLevel)

	for i := randLevel; i >= 0; i-- {
		newElem.levels[i] = prevs[i].levels[i]
		prevs[i].levels[i] = newElem
	}
	return nil
}

func (list *SkipList) Get(key []byte) (e *codec.Entry) {
	list.RLock()
	defer list.RUnlock()

	prev := list.header
	keyScore := list.calcScore(key)

	for i := list.maxLevel; i >= 0; i-- {
		for next := prev.levels[i]; next != nil; next = prev.levels[i] {
			v := list.compare(keyScore, key, next)
			if v <= 0 {
				if v == 0 {
					return next.entry
				}
				break
			}
			prev = next
		}
	}
	return
}

// 取key的前8位做hash
func (list *SkipList) calcScore(key []byte) (score float64) {
	var hash uint64
	l := len(key)

	if l > 8 {
		l = 8
	}

	for i := 0; i < l; i++ {
		shift := uint(64 - 8 - i*8)
		hash |= uint64(key[i]) << shift
	}
	score = float64(hash)
	return
}

func (list *SkipList) compare(score float64, key []byte, next *Element) int {
	if score == next.score {
		return bytes.Compare(key, next.entry.Key)
	}
	if score < next.score {
		return -1
	}
	return 1
}

func (list *SkipList) randLevel() int {
	i := 1
	for ; ; i++ {
		if rand.Intn(2) == 0 {
			return i
		}
	}
}
