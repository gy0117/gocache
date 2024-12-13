package codec

type Entry struct {
	Key []byte
	Val []byte
}

func NewEntry(key, val []byte) *Entry {
	return &Entry{
		Key: key,
		Val: val,
	}
}
