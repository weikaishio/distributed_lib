package bloom_filter

import "encoding/binary"

func (b *Bloom) AddString(item string) error {
	return b.AddItem([]byte(item))
}
func (b *Bloom) AddUint32(item uint32) error {
	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data, item)
	return b.AddItem(data)
}
func (b *Bloom) AddUint64(item uint64) error {
	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data, item)
	return b.AddItem(data)
}

func (b *Bloom) AddItem(item []byte) (err error) {
	b.Add(item)
	err = b.Sync()
	return
}
func (b *Bloom) AddItemWithoutSync(item []byte) {
	b.Add(item)
}
