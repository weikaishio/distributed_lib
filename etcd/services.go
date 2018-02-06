package etcd

import (
	"context"

	"fmt"
	"time"

	"sync"

	"github.com/coreos/etcd/clientv3"
	etcdnaming "github.com/coreos/etcd/clientv3/naming"
	"github.com/coreos/etcd/etcdserver/api/v3rpc/rpctypes"
	"github.com/mkideal/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/naming"
)

type Discovery struct {
	cli        *clientv3.Client
	resolver   *etcdnaming.GRPCResolver //grpc所用
	stopSignal chan struct{}
	serviceMap map[string]time.Duration
	lockSvrMap sync.RWMutex
}

const (
	PREFIX = "ServicePref"
)

func NewDiscovery(endpoints []string, username, password string) (*Discovery, error) {
	cliCfg := clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: DIALTIMEOUT,
		Username:    username,
		Password:    password,
	}
	cli, err := clientv3.New(cliCfg)
	if err != nil {
		return nil, err
	}
	return &Discovery{
		cli:        cli,
		resolver:   &etcdnaming.GRPCResolver{Client: cli},
		stopSignal: make(chan struct{}),
		serviceMap: make(map[string]time.Duration),
	}, nil
}
func (c *Discovery) Register(serviceName string, host string, port int, interval time.Duration, metaData string) {
	//c.resolver.Update(context.TODO(), key, naming.Update{})
	serviceValue := fmt.Sprintf("%s:%d", host, port)
	serviceKey := fmt.Sprintf("/%s/%s", PREFIX, serviceName)
	c.lockSvrMap.Lock()
	c.serviceMap[serviceValue] = interval
	c.lockSvrMap.Unlock()
	// get endpoints for register dial address
	go func() {
		// invoke self-register with ticker
		ticker := time.NewTicker(interval)
		for {
			var (
				opOption clientv3.OpOption
				//lease    clientv3.Lease
				err error
			)
			// minimum lease TTL is ttl-second
			//lease = clientv3.NewLease(c.cli)
			//resp, err := lease.Grant(context.TODO(), int64(interval/time.Second))
			//if err != nil {
			//	log.Error("")
			//	return
			//}
			//opOption = clientv3.WithLease(resp.ID)
			// should get first, if not exist, set it
			_, err = c.cli.Get(context.Background(), serviceKey)
			if err != nil {
				if err == rpctypes.ErrKeyNotFound {
					if err := c.resolver.Update(context.TODO(), serviceKey, naming.Update{
						Op:   naming.Add,
						Addr: serviceValue,
					}, opOption); err != nil {
						log.Error("grpclb: set service '%s' with ttl to etcd3 failed: %s", serviceKey, err.Error())
					} else {
						log.Info("c.resolver.Update1 :%s,addr:%s", serviceKey, serviceValue)
					}
				} else {
					log.Error("grpclb: service '%s' connect to etcd3 failed: %s", serviceKey, err.Error())
				}
			} else {
				//refresh set to true for not notifying the watcher
				//if _, err := c.cli.Put(context.Background(), serviceKey+"/"+serviceValue, serviceValue, clientv3.WithLease(resp.ID)); err != nil {
				//	log.Error("grpclb: refresh service '%s' with ttl to etcd3 failed: %s", serviceKey, err.Error())
				//}
				if err := c.resolver.Update(context.TODO(), serviceKey, naming.Update{
					Op:       naming.Add,
					Addr:     serviceValue,
					Metadata: metaData,
				}); err != nil {
					log.Error("grpclb: refresh service '%s' with ttl to etcd3 failed: %s", serviceKey, err.Error())
				} else {
					log.Info("c.resolver.Update2 :%s,addr:%s", serviceKey, serviceValue)
				}
			}
			select {
			case <-c.stopSignal:
				log.Warn("rev stopSignal")
				return
			case <-ticker.C:
			}
		}
	}()
}

func (c *Discovery) UnRegister(serviceName string, host string, port int) error {
	interval := 10 * time.Second
	serviceValue := fmt.Sprintf("%s:%d", host, port)
	c.lockSvrMap.RLock()
	if v, ok := c.serviceMap[serviceValue]; ok {
		interval = v
	}
	c.lockSvrMap.RUnlock()
	select {
	case c.stopSignal <- struct{}{}:
		log.Info("stop Register ok to stop")
	case <-time.Tick(interval / 2): //Register not on running can exec next
		log.Info("time to stop")
	}
	c.stopSignal = make(chan struct{}, 1) // just a hack to avoid multi UnRegister deadlock
	var err error
	serviceKey := fmt.Sprintf("/%s/%s", PREFIX, serviceName)
	//if _, err := c.cli.Delete(context.Background(), serviceKey+"/"+serviceValue); err != nil {
	//	log.Error("grpclb: deregister '%s' failed: %s", serviceKey, err.Error())
	//} else {
	//	log.Info("grpclb: deregister '%s' ok.", serviceKey)
	//}
	err = c.resolver.Update(context.TODO(), serviceKey, naming.Update{
		Op:   naming.Delete,
		Addr: serviceValue,
	})
	return err
}

func (c *Discovery) NewBalancer() grpc.Balancer {
	return grpc.RoundRobin(c.resolver)
}
func (c *Discovery) Dial(serviceName string) (*grpc.ClientConn, error) {
	b := c.NewBalancer()
	service := fmt.Sprintf("/%s/%s", PREFIX, serviceName)
	// must add DialOption grpc.WithBlock, or it will rpc error: code = Unavailable desc = there is no address available
	conn, err := grpc.Dial(service, grpc.WithBalancer(b), grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(DIALTIMEOUT))
	if err != nil {
		log.Error("Discovery service:%s,err:%v", service, err)
	}
	return conn, err
}
func (c *Discovery) WatchService(serviceName string) error {
	serviceKey := fmt.Sprintf("/%s/%s", PREFIX, serviceName)
	watch, err := c.resolver.Resolve(serviceKey)
	if err != nil {
		return err
	}
	for {
		upAry, _ := watch.Next()
		for _, up := range upAry {
			log.Info("up:%v", up)
		}
	}
}
