package etcd

import (
	"testing"

	"fmt"

	"github.com/mkideal/log"
)

var (
	cfgSvr *ConfigSvr
	//endpoints  = []string{"127.0.0.1:2379"}
	endpoints  = []string{"52.187.8.211:2379"}
	discovery  *Discovery
	serviceKey = fmt.Sprintf("/%s/%s", PREFIX, key)
)

/*
http://blog.csdn.net/zstack_org/article/details/54924651
/tmp/etcd-download-test/etcdctl ls / --recursive
etcdctl -o extended get xxx
etcdctl mkdir /example
etcdctl update /example/key turtles
etcdctl mkdir /here/you/go --ttl 120

etcdctl rm /a --recursive
*/
const (
	prefix = "/service_healthjob"
	key    = "WatchSvr"
)

func init() {
	var err error
	cfgSvr, err = NewConfigSvr(endpoints)
	if err != nil {
		log.Fatal("NewConfigSvr err:%v", err)
	}
	discovery, err = NewDiscovery(endpoints)
	if err != nil {
		log.Fatal("NewDiscovery err:%v", err)
	}
}

func TestConfigSvr_MultiSetKV(t *testing.T) {
	kv := make(map[string]string)

	kv[prefix+"/log:level"] = "level"
	kv[prefix+"/log:path"] = "path"
	kv[prefix+"/job:runrule"] = "121"
	kv[prefix+"/job:port"] = "1"
	kv[prefix+"/job:isProduction"] = "1"

	kv[prefix+"/mysql:sg:driver"] = "1"
	kv[prefix+"/mysql:sg:host"] = "2"
	kv[prefix+"/mysql:sg:database"] = "2"
	kv[prefix+"/mysql:sg:username"] = "2"
	kv[prefix+"/mysql:sg:password"] = "2"

	kv[prefix+"/redis:sg:host"] = "2"
	kv[prefix+"/redis:sg:password"] = "2"
	kv[prefix+"/redis:sg:database"] = "2"
	kv[prefix+"/redis:usw:host"] = "2"
	kv[prefix+"/redis:usw:password"] = "password"
	kv[prefix+"/redis:usw:database"] = "2"

	cfgSvr.MultiSetKV(kv)
}

//func TestConfigSvr_SetKV(t *testing.T) {
//	key := "/service_healthjob/redis:sg:host"
//	go func() {
//		cfgSvr.Watch(key)
//	}()
//	cfgSvr.SetKV(key, "value"+time.Now().Format("20060102150405"))
//
//	kvs, _ := cfgSvr.GetKey(key)
//	if len(kvs) > 0 {
//		for _, v := range kvs {
//			t.Logf("k:%s,v:%s", v.Key, v.Value)
//		}
//	}
//	time.Sleep(10 * time.Second)
//}

//func TestDiscovery_Register(t *testing.T) {
//	t.Logf("serviceKey:%s", serviceKey)
//	rand.Seed(time.Now().Unix())
//	port := 678 //rand.Intn(1000)
//	discovery.Register(key, "127.0.0.1", port, 15*time.Second, "test")
//	stop := make(chan struct{})
//	go func() { //Register 的key相当于 fmt.Sprintf("/%s/%s", PREFIX, key)+addr
//		cfgSvr.Watch(serviceKey + "/127.0.0.1:" + strconv.Itoa(port))
//	}()
//	go func() {
//		discovery.Dial(serviceKey)
//		time.Sleep(5 * time.Second)
//		log.Info("UnRegister:" + "/127.0.0.1:" + strconv.Itoa(port))
//		discovery.UnRegister(key, "127.0.0.1", port)
//		log.Info("UnRegister finish")
//
//		stop <- struct{}{}
//	}()
//	func() {
//		log.Info("begin watch")
//		watch, err := discovery.resolver.Resolve(serviceKey)
//		defer watch.Close()
//		t.Logf("watch:%v,err:%v", watch, err)
//		for {
//			upAry, _ := watch.Next()
//			for _, up := range upAry {
//				log.Info("up:%v", up)
//			}
//			select {
//			case <-stop:
//				return
//			case <-time.Tick(time.Millisecond):
//			}
//		}
//	}()
//	log.Info("stop")
//}
