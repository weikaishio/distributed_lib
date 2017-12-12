package etcd

import (
	"time"

	"context"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/etcdserver/api/v3rpc/rpctypes"
	"github.com/coreos/etcd/mvcc/mvccpb"

	"github.com/mkideal/log"
)

type ConfigSvr struct {
	cli *clientv3.Client
}

const (
	DIALTIMEOUT = 15 * time.Second
	SETTIMEOUT  = 5 * time.Second
)

func NewConfigSvr(endpoints []string) (*ConfigSvr, error) {
	cliCfg := clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: DIALTIMEOUT,
	}
	cli, err := clientv3.New(cliCfg)
	if err != nil {
		return nil, err
	}
	return &ConfigSvr{
		cli: cli,
	}, nil
}
func (c *ConfigSvr) SetKV(key, value string) {
	ctx, cancel := context.WithTimeout(context.Background(), SETTIMEOUT)
	resp, err := c.cli.Put(ctx, key, value)
	cancel()
	if err != nil {
		switch err {
		case context.Canceled:
			log.Error("ctx is canceled by another routine: %v", err)
		case context.DeadlineExceeded:
			log.Error("ctx is attached with a deadline is exceeded: %v", err)
		case rpctypes.ErrEmptyKey:
			log.Error("client-side error: %v", err)
		default:
			log.Error("bad cluster endpoints, which are not etcd servers: %v", err)
		}
	} else {
		log.Info("resp:%v", resp)
	}
}

/*
Conf来监听对应的key 做后续操作即可，此处暂时无用
*/
func (c *ConfigSvr) Watch(key string) {
	watchCH := c.cli.Watch(context.Background(), key)
	for wresp := range watchCH {
		for _, ev := range wresp.Events {
			switch ev.Type {
			case mvccpb.PUT:
				log.Info("PUT:%v", string(ev.Kv.Value))
				//return []*naming.Update{{Op: naming.Add, Addr: string(ev.Kv.Value)}}, nil
			case mvccpb.DELETE:
				log.Info("DELETE:%v", string(ev.Kv.Value))
				//return []*naming.Update{{Op: naming.Delete, Addr: string(ev.Kv.Value)}}, nil
			}
		}
	}
	//c.cli.Close()
}
