package bloom_filter

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"github.com/willf/bloom"
	"io/ioutil"
)

type Bloom struct {
	*bloom.BloomFilter
	key           string
	isGZip        bool
	loadFromStore func(key string) ([]byte, error)
	sync2Store    func(key string, data []byte) error
}

func NewBloom(m, k uint,
	isGZip bool,
	key string,
	loadFromStore func(key string) ([]byte, error),
	sync2Store func(key string, data []byte) error) *Bloom {
	return &Bloom{
		BloomFilter:   bloom.New(m, k),
		isGZip:        isGZip,
		key:           key,
		loadFromStore: loadFromStore,
		sync2Store:    sync2Store,
	}
}
func (b *Bloom) Load() (err error) {
	if b.loadFromStore == nil {
		return errors.New("Bloom.loadFromStore is invalid")
	}
	if b.key == "" {
		return errors.New("Bloom.key is invalid")
	}
	var data []byte
	data, err = b.loadFromStore(b.key)
	if err != nil {
		err = fmt.Errorf("loadFromStore fail err:%v", err)
		return
	}
	if data != nil &&len(data)>0{
		if b.isGZip {
			var unzipData []byte
			unzipData, err = b.unGZip(data)
			if err != nil {
				err = fmt.Errorf("data ungzip fail err:%v", err)
				return
			}
			err = b.UnmarshalJSON(unzipData)
		} else {
			err = b.UnmarshalJSON(data)
		}
	}
	return
}
func (b *Bloom) unGZip(data []byte) (unzipData []byte, err error) {
	var reader *gzip.Reader
	reader, err = gzip.NewReader(bytes.NewReader(data))
	if reader != nil {
		defer func() {
			_ = reader.Close()
		}()
	}
	if err != nil {
		return
	}
	unzipData, err = ioutil.ReadAll(reader)
	if err != nil {
		return
	}
	return
}

func (b *Bloom) Sync() (err error) {
	if b.sync2Store != nil {
		var data []byte
		data, err = b.MarshalJSON()
		if err != nil {
			err = fmt.Errorf("failed to marshal bloom data err: %v", err)
			return
		}

		if b.isGZip {
			var buf bytes.Buffer
			gw := gzip.NewWriter(&buf)
			_, _ = gw.Write(data)
			_ = gw.Close()

			data, err = ioutil.ReadAll(&buf)
			if err != nil {
				err = fmt.Errorf("failed to gzip bloom data err: %v", err)
				return
			}
		}
		err = b.sync2Store(b.key, data)
		if err != nil {
			err = fmt.Errorf("failed to sync2Store err: %v", err)
			return
		}
	}
	return
}
