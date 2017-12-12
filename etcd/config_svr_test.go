package etcd

import (
	"math/rand"
	"testing"

	"fmt"

	"time"

	"strconv"

	"github.com/mkideal/log"
)

var (
	cfgSvr     *ConfigSvr
	endpoints  = []string{"127.0.0.1:2379"}
	discovery  *Discovery
	serviceKey = fmt.Sprintf("/%s/%s", PREFIX, key)
)

const (
	key = "WatchSvr"
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

//func TestConfigSvr_SetKV(t *testing.T) {
//	go func() {
//		cfgSvr.Watch(key)
//	}()
//	for i := 0; i < 10; i++ {
//		cfgSvr.SetKV(key, "value"+time.Now().Format("20060102150405"))
//
//		time.Sleep(time.Second)
//	}
//	time.Sleep(10 * time.Second)
//}
func TestDiscovery_Register(t *testing.T) {
	t.Logf("serviceKey:%s", serviceKey)
	rand.Seed(time.Now().Unix())
	port := 678 //rand.Intn(1000)
	discovery.Register(key, "127.0.0.1", port, 15*time.Second, "test")
	stop := make(chan struct{})
	go func() { //Register 的key相当于 fmt.Sprintf("/%s/%s", PREFIX, key)+addr
		cfgSvr.Watch(serviceKey + "/127.0.0.1:" + strconv.Itoa(port))
	}()
	go func() {
		discovery.Dial(serviceKey)
		time.Sleep(5 * time.Second)
		log.Info("UnRegister:" + "/127.0.0.1:" + strconv.Itoa(port))
		discovery.UnRegister(key, "127.0.0.1", port)
		log.Info("UnRegister finish")

		stop <- struct{}{}
	}()
	func() {
		log.Info("begin watch")
		watch, err := discovery.resolver.Resolve(serviceKey)
		defer watch.Close()
		t.Logf("watch:%v,err:%v", watch, err)
		for {
			upAry, _ := watch.Next()
			for _, up := range upAry {
				log.Info("up:%v", up)
			}
			select {
			case <-stop:
				return
			case <-time.Tick(time.Millisecond):
			}
		}
	}()
	log.Info("stop")
}
