package buffer

import (
	"sync"
	"time"

	"github.com/mkideal/log"
)

const (
	maxWaitTime = time.Second
)

type Pool struct {
	pool       *sync.Pool
	collection chan *Buffer
}

func NewPool(initSize, maxSize int) *Pool {
	if maxSize < initSize {
		maxSize = initSize
	}
	bp := &Pool{
		pool:       &sync.Pool{New: func() interface{} { return NewBuffer() }},
		collection: make(chan *Buffer, maxSize),
	}
	for i := 0; i < initSize; i++ {
		bp.pool.Put(bp.pool.New())
	}
	go bp.run()
	return bp
}

func (bp *Pool) Get() (buf *Buffer, fromPool bool) {
	bufObj := bp.pool.Get()
	if bufObj == nil {
		buf = NewBuffer()
	} else {
		fromPool = true
		buf = bufObj.(*Buffer)
		buf.Reset()
	}
	return
}

func (bp *Pool) Put(buf *Buffer, fromPool bool) {
	if fromPool && buf != nil {
		select {
		case bp.collection <- buf:
		default:
			log.Warn("put buffer fail")
		}
	}
}

func (bp *Pool) run() {
	for buf := range bp.collection {
		if buf.Wait(maxWaitTime) {
			bp.pool.Put(buf)
		}
	}
}
