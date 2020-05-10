package bloom_filter

import "testing"

var (
	b *Bloom
)

func init() {
	b = NewBloom(500000, 5, false, "test", nil, nil)
	_ = b.Load()
}

func TestBloom_AddUint64(t *testing.T) {
	var i uint64
	for i = 1; i < 20000; i++ {
		_ = b.AddUint64(i)
	}
}

func TestBloom_ContainsUint64(t *testing.T) {
	var i uint64
	failCount := 0
	for i = 1; i < 40000; i++ {
		has := b.ContainsUint64(i)
		if has && i > 20000 {
			failCount++
		}
	}
	t.Logf("TestBloom_ContainsUint64 failCount:%v", failCount)
}
