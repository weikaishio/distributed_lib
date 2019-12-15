package lru_cache

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	keyTest = "key_test"
	keyData = "data_test"
	setFunc = func(key string) interface{} {
		return keyData
	}
)

func TestInProcCache_SetValFunc(t *testing.T) {
	InProcCacheSharedInstance().SetValFunc(keyTest, setFunc)
	data, has := InProcCacheSharedInstance().Get(keyTest)
	assert.Equal(t, has, true)
	assert.Equal(t, data, keyData)
}
func TestInProcCache_Set(t *testing.T) {
	InProcCacheSharedInstance().SetLimitCount(2)
	InProcCacheSharedInstance().Set(keyTest, keyData)
	data, has := InProcCacheSharedInstance().Get(keyTest)
	assert.Equal(t, has, true)
	assert.Equal(t, data, keyData)

	for i := 0; i < 3; i++ {
		InProcCacheSharedInstance().Set(fmt.Sprintf("%s:%d", keyTest, i), fmt.Sprintf("%s:%d", keyData, i))
	}
	for i := 0; i < 3; i++ {
		data, has := InProcCacheSharedInstance().Get(fmt.Sprintf("%s:%d", keyTest, i))
		t.Logf("data:%v,has:%v", data, has)
	}
}
