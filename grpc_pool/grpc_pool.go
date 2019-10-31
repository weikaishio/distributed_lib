package grpc_pool

import (
	"sync"
	"time"

	"github.com/mkideal/log"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

const (
	pool_min_size     = 3
	pool_max_size     = 66
	pool_conn_timeout = 600 * time.Second
)

type Connection interface {
	Close() error
}
type PoolConn struct {
	Connection
	t time.Time
}
type RPCPool struct {
	poolChan chan *PoolConn
	mu       sync.RWMutex
	makeConn func() (Connection, error)
}

func NewRPCPool(makeGRPCConn func() (Connection, error)) *RPCPool {
	return &RPCPool{
		poolChan: make(chan *PoolConn, pool_max_size),
		mu:       sync.RWMutex{},
		makeConn: makeGRPCConn,
	}
}

func (p *RPCPool) InitPool() {
	for i := 0; i < pool_min_size; i++ {
		if len(p.poolChan) == pool_max_size {
			break
		}
		conn := p.CreateConn()
		if conn != nil {
			p.poolChan <- &PoolConn{conn, time.Now()}
		}
	}
}

func (p *RPCPool) getPool() chan *PoolConn {
	p.mu.RLock()
	conns := p.poolChan
	p.mu.RUnlock()
	return conns
}

func (p *RPCPool) CreateConn() Connection {
	conn, err := p.makeConn()
	if err != nil {
		log.Error("RPCPool CreateConn makeConn err:%v", err)
		return nil
	} else {
		return conn
	}
}
func (p *RPCPool) Borrow() Connection {
	conns := p.getPool()
	if conns == nil {
		return nil
	}
	for {
		select {
		case conn := <-conns:
			if conn.t.Add(pool_conn_timeout).Before(time.Now()) {
				log.Trace("RPCPool Borrow conn is time of arrival close.")
				_ = conn.Close()
				continue
			} else {
				grpcConn := conn.Connection.(*grpc.ClientConn)
				if grpcConn != nil {
					if grpcConn.GetState() != connectivity.Ready {
						log.Warn("RPCPool Borrow conn.state not in ready, it state is:%s", grpcConn.GetState().String())
						switch grpcConn.GetState() {
						case connectivity.Idle:
						case connectivity.Connecting:
						case connectivity.TransientFailure:
							fallthrough
						case connectivity.Shutdown:
							fallthrough
						default:
							_ = conn.Close()
						}
						continue
					} else {
						return conn.Connection
					}
				}else{
					continue
				}
			}
		case <-time.After(time.Second):
			if len(conns) < pool_max_size {
				log.Trace("Borrow,len(p.poolChan):%d < pool_max_size", len(conns))
				conn := p.CreateConn()
				if conn != nil {
					log.Trace("RPCPool Borrow AddConn2Pool,  pool's len is %d after created conn", len(conns))
					return conn
				}
			}
		}
	}
}
func (p *RPCPool) Return(conn Connection) {
	if conn == nil || p.poolChan == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	select {
	case p.poolChan <- &PoolConn{conn, time.Now()}:
	default:
		log.Warn("RPCPool: conn recycled fail, It may be full")
		_ = conn.Close()
	}
	log.Trace("RPCPool: conn has been recycled conn, now it len:%d", len(p.poolChan))
}

func (p *RPCPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for cli := range p.poolChan {
		_ = cli.Close()
	}
	close(p.poolChan)
	log.Trace("RPCPool Closed")
}
