package bloom_filter

import "encoding/binary"

func (b *Bloom) ContainsString(item string) bool {
	return b.Contains([]byte(item))
}
func (b *Bloom) ContainsUint32(item uint32) bool {
	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data, item)
	return b.Contains(data)
}
func (b *Bloom) ContainsUint64(item uint64) bool {
	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data, item)
	return b.Contains(data)
}
func (b *Bloom) Contains(item []byte) bool {
	return b.Test(item)
}
