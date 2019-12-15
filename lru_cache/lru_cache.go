package lru_cache

import (
	"sync"
	"sync/atomic"
)

type LRUCache struct {
	caches       sync.Map
	limitCount   int32
	currentCount int32
	headNode     *LinkedListNode
	endNode      *LinkedListNode
}

func NewLRUCache() *LRUCache {
	return &LRUCache{
		caches:       sync.Map{},
		limitCount:   10000,
		currentCount: 0,
		headNode:     nil,
		endNode:      nil,
	}
}
func (c *LRUCache) PutNode(pNode *LinkedListNode) {
	if c.endNode != nil {
		c.endNode.Next = pNode
		pNode.Pre = c.endNode
		pNode.Next = nil
	}
	c.endNode = pNode

	if c.headNode == nil {
		c.headNode = pNode
	}
}
func (c *LRUCache) RemoveNode(pNode *LinkedListNode) {
	if c.endNode == pNode {
		c.endNode = c.endNode.Pre
	} else if c.headNode == pNode {
		c.headNode = c.headNode.Next
	} else {
		if pNode.Next != nil {
			pNode.Next.Pre = pNode.Pre
		}
		if pNode.Pre != nil {
			pNode.Pre.Next = pNode.Next
		}
	}
}

func (c *LRUCache) SetLimitCount(count int32) {
	c.limitCount = count
}

func (c *LRUCache) Get(key string) (interface{}, bool) {
	curData, has := c.caches.Load(key)
	if has {
		pNode := curData.(*LinkedListNode)
		if pNode != c.endNode {
			c.RemoveNode(pNode)
			c.PutNode(pNode)
		}
		return pNode.Data, has
	}
	c.caches.Range(func(key, value interface{}) bool {
		
	})
	return nil, false
}
func (c *LRUCache) Set(key string, data interface{}) {
	var pNode *LinkedListNode
	curData, has := c.caches.Load(key)
	if has {
		pNode = curData.(*LinkedListNode)
		c.RemoveNode(pNode)
	} else {
		pNode = &LinkedListNode{
			Key:  key,
			Data: data,
		}
		atomic.AddInt32(&c.currentCount, 1)
		c.VerifyIsOverLimit()
	}
	c.caches.Store(key, pNode)
	c.PutNode(pNode)
}
func (c *LRUCache) VerifyIsOverLimit() {
	currentCount := atomic.LoadInt32(&c.currentCount)
	if currentCount > c.limitCount {
		c.Del(c.headNode.Key)
	}
}
func (c *LRUCache) Del(key interface{}) {
	curData, has := c.caches.Load(key)
	if has {
		pNode := curData.(*LinkedListNode)
		c.RemoveNode(pNode)
		c.caches.Delete(key)
		atomic.AddInt32(&c.currentCount, -1)
	}
}
