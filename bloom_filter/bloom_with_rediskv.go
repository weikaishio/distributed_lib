package bloom_filter

import (
	"github.com/go-redis/redis"
	"time"
)

const (
	m = 2 << 18
	k = 4
)

type BloomWithRedis struct {
	size, hashCount uint
	isGZip          bool
	loadFromStore   func(key string) ([]byte, error)
	sync2Store      func(key string, data []byte) error
}

func NewBloomWithRedis(redisCli redis.Cmdable) (b *BloomWithRedis, err error) {
	return NewBloomWithRedisWithMAndK(redisCli, m, k, false)
}
func NewBloomWithRedisWithMAndK(redisCli redis.Cmdable,
	size, hashCount uint,
	isGZip bool) (b *BloomWithRedis, err error) {
	loadFromStore := func(key string) ([]byte, error) {
		data,err:= redisCli.Get(key).Bytes()
		if err==redis.Nil{
			err=nil
		}
		return data,err
	}
	sync2Store := func(key string, data []byte) error {
		_,err:= redisCli.Set(key, data,-1*time.Second).Result()
		return err
	}
	b = &BloomWithRedis{
		size:          size,
		hashCount:     hashCount,
		isGZip:        isGZip,
		loadFromStore: loadFromStore,
		sync2Store:    sync2Store,
	}
	return
}
func (b *BloomWithRedis) Contains(key string, items ...interface{}) (contains []interface{}, err error) {
	bloom := NewBloom(b.size, b.hashCount, b.isGZip, key, b.loadFromStore, b.sync2Store)
	err = bloom.Load()
	if err != nil {
		return
	}

	for _, item := range items {
		switch typ := item.(type) {
		case uint64:
			if bloom.ContainsUint64(typ) {
				contains = append(contains, item)
			}
		case uint32:
			if bloom.ContainsUint32(typ) {
				contains = append(contains, item)
			}
		case string:
			if bloom.ContainsString(typ) {
				contains = append(contains, item)
			}
		}
	}
	return
}
func (b *BloomWithRedis) AddAndContains(key string, items ...interface{}) (notContains []interface{}, err error) {
	bloom := NewBloom(b.size, b.hashCount, b.isGZip, key, b.loadFromStore, b.sync2Store)
	err = bloom.Load()
	if err != nil {
		return
	}

	for _, item := range items {
		switch typ := item.(type) {
		case uint64:
			if !bloom.ContainsUint64(typ) {
				notContains = append(notContains, item)
				err = bloom.AddUint64(typ)
				if err != nil {
					return
				}
			}
		case uint32:
			if !bloom.ContainsUint32(typ) {
				notContains = append(notContains, item)
				err = bloom.AddUint32(typ)
				if err != nil {
					return
				}
			}
		case string:
			if !bloom.ContainsString(typ) {
				notContains = append(notContains, item)
				err = bloom.AddString(typ)
				if err != nil {
					return
				}
			}
		}
	}
	err = bloom.Sync()
	return
}
