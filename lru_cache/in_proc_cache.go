package lru_cache

import "sync"

type InProcCache struct {
	cache    *LRUCache
	funcLock sync.RWMutex
	funcMap  map[string]func(string) interface{}
}

var (
	inProcCacheInstance *InProcCache
	once                      sync.Once
)

func InProcCacheSharedInstance() *InProcCache {
	once.Do(func() {
		inProcCacheInstance = &InProcCache{
			cache:    NewLRUCache(),
			funcLock: sync.RWMutex{},
			funcMap:  make(map[string]func(string) interface{}),
		}
	})
	return inProcCacheInstance
}
func (c *InProcCache) SetLimitCount(count int32) {
	c.cache.SetLimitCount(count)
}
func (c *InProcCache) SetValFunc(key string, f func(string) interface{}) {
	c.funcLock.Lock()
	defer c.funcLock.Unlock()
	c.funcMap[key] = f
}
//func (c *InProcCache) SetInitCacheFunc(f func(key, value interface{}) bool){
//
//}

func (c *InProcCache) Get(key string) (interface{}, bool) {
	data, has := c.cache.Get(key)
	if has {
		return data, has
	}
	c.funcLock.RLock()
	setFunc, has := c.funcMap[key]
	c.funcLock.RUnlock()
	if has {
		newData := setFunc(key)
		c.Set(key, newData)
		return newData, true
	}
	return nil, false
}
func (c *InProcCache) Set(key string, data interface{}) {
	c.cache.Set(key, data)
}
