package utils

import (
	"fmt"
	"github.com/gy0117/gocache/pkg/utils/codec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

func TestSkipList_compare(t *testing.T) {
	list := &SkipList{
		header: nil,
		rand:   nil,
	}
	b1 := []byte("1")
	b2 := []byte("2")
	entry1 := &codec.Entry{
		Key: b1,
		Val: b2,
	}
	b1Score := list.calcScore(b1)
	b2Score := list.calcScore(b2)

	element := &Element{
		levels: nil,
		entry:  entry1,
		score:  b2Score,
	}
	assert.Equal(t, list.compare(b1Score, b1, element), -1)
}

func TestSkipList_curd(t *testing.T) {
	list := NewSkipList()

	// put && get
	entry1 := &codec.Entry{
		Key: []byte("key1"),
		Val: []byte("val1"),
	}
	assert.Nil(t, list.Put(entry1))
	assert.Equal(t, entry1.Val, list.Get(entry1.Key).Val)

	entry2 := &codec.Entry{
		Key: []byte("key2"),
		Val: []byte("val2"),
	}
	assert.Nil(t, list.Put(entry2))
	assert.Equal(t, entry2.Val, list.Get(entry2.Key).Val)

	// get a not-exist entry
	assert.Nil(t, list.Get([]byte("kk")))

	// update entry
	updateEntry2 := &codec.Entry{
		Key: []byte("key1"),
		Val: []byte("val123"),
	}
	assert.Nil(t, list.Put(updateEntry2))
	assert.Equal(t, updateEntry2.Val, list.Get(updateEntry2.Key).Val)
}

func TestSkipList_concurrent(t *testing.T) {
	n := 10000
	l := NewSkipList()
	var wg sync.WaitGroup
	key := func(i int) []byte {
		return []byte(fmt.Sprintf("%05d", i))
	}
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			assert.Nil(t, l.Put(&codec.Entry{Key: key(i), Val: key(i)}))
		}(i)
	}
	wg.Wait()

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			v := l.Get(key(i))
			if v != nil {
				require.EqualValues(t, key(i), v.Val)
				return
			}
			require.Nil(t, v)
		}(i)
	}
	wg.Wait()
}

func Benchmark_SkipList_crud(b *testing.B) {
	list := NewSkipList()
	key, val := "", ""
	for i := 0; i < b.N; i++ {
		key, val = fmt.Sprintf("key%d", i), fmt.Sprintf("val%d", i)
		entry := &codec.Entry{
			Key: []byte(key),
			Val: []byte(val),
		}
		res := list.Put(entry)
		assert.Equal(b, res, nil)

		getVal := list.Get([]byte(key))
		assert.Equal(b, getVal.Val, []byte(val))
	}
}
