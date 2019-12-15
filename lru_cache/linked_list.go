package lru_cache

type LinkedListNode struct {
	Key  string
	Data interface{}
	Pre  *LinkedListNode
	Next *LinkedListNode
}
