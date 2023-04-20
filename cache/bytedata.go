package cache

// 缓存数据
type ByteData struct {
	data []byte
}

func (db ByteData) Len() int {
	return len(db.data)
}

func (db ByteData) ByteSlice() []byte {
	return cloneBytes(db.data)
}

func cloneBytes(in []byte) []byte {
	out := make([]byte, len(in))
	copy(out, in)
	return out
}

func (db ByteData) String() string {
	return string(db.data)
}
