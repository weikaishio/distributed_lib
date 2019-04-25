package redis_pubsub

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-redis/redis"
	"github.com/mkideal/log"
)

type RdsPubSubMsg struct {
	redisClient *redis.Client
	subCli      *redis.PubSub
	updateLock  sync.RWMutex
	newMsgRev   map[string]func(msg interface{})
	lock        sync.RWMutex
	running     int32
	newChannel  chan struct{}
}

var (
	rdsPubSubClient  *RdsPubSubMsg
	once             sync.Once
	ERR_RedisNotInit = errors.New("redis not init")
)

var (
	rdsPubSubMsgClient *RdsPubSubMsg
	onceMsg            sync.Once
)

func SharedRdsSubscribMsgInstance() *RdsPubSubMsg {
	onceMsg.Do(func() {
		rdsPubSubMsgClient = &RdsPubSubMsg{
			newMsgRev:  make(map[string]func(msg interface{})),
			newChannel: make(chan struct{}),
		}
	})
	return rdsPubSubMsgClient
}

func (r *RdsPubSubMsg) AddSubscribe(channel string, onRevMsg func(msg interface{})) {
	r.updateLock.Lock()
	defer r.updateLock.Unlock()
	r.newMsgRev[channel] = onRevMsg
	go func() {
		r.newChannel <- struct{}{}
	}()
}

func (r *RdsPubSubMsg) Publish(channel string, msg interface{}) error {
	if r.redisClient != nil {
		_, err := r.redisClient.Publish(channel, msg).Result()
		return err
	} else {
		return ERR_RedisNotInit
	}
}

func (r *RdsPubSubMsg) Quit() {
	log.Info("RdsPubSub ready quit")
	atomic.SwapInt32(&r.running, 0)
	log.Info("RdsPubSubquit ok")
}
func (r *RdsPubSubMsg) IsRunning() bool {
	return atomic.LoadInt32(&r.running) != 0
}
func (r *RdsPubSubMsg) Set(rdsCli *redis.Client) {
	if r.redisClient == nil {
		r.redisClient = rdsCli
	}
}
func (r *RdsPubSubMsg) StartSubscription() {
	if !atomic.CompareAndSwapInt32(&r.running, 0, 1) {
		return
	}

	log.Info("StartSubscription")
	defer atomic.CompareAndSwapInt32(&r.running, 1, 0)
	var channels []string
	for k, _ := range r.newMsgRev {
		channels = append(channels, k)
	}

	sleepSecond := 3
	for r.IsRunning() {
		select {
		case <-time.After(1 * time.Second):
			if r.redisClient == nil {
				log.Warn("StartSubscription rdsPubSubClient.redisClient is nil")
				break
			}
			subCli := r.redisClient.PSubscribe(channels...)
			for r.IsRunning() {
				isOpen, _ := r.subscription(subCli, sleepSecond, channels)

				if !isOpen {
					subCli = r.redisClient.PSubscribe(channels...)
					log.Error("StartSubscription sub.Channel() isClose, u.renewSubClient")
					time.Sleep(time.Duration(sleepSecond) * time.Second)
					sleepSecond += sleepSecond
				}
			}
		}
	}
	log.Info("quit StartSubscription")
}
func (r *RdsPubSubMsg) subscription(subCli *redis.PubSub, sleepSecond int, channels []string) (isOpen bool, err error) {
	defer func() {
		if e := recover(); e != nil {
			buf := make([]byte, 1<<16)
			buf = buf[:runtime.Stack(buf, true)]
			switch typ := e.(type) {
			case error:
				err = typ
			case string:
				err = errors.New(typ)
			default:
				err = fmt.Errorf("%v", typ)
			}
			log.Error("==== STACK TRACE BEGIN ====\npanic: %v\n%s\n===== STACK TRACE END =====", err, string(buf))
		}
	}()
	select {
	case msg, isOpen := <-subCli.Channel():
		if isOpen {
			r.updateLock.RLock()
			onRevMsg, has := r.newMsgRev[msg.Channel]
			r.updateLock.RUnlock()
			if has {
				onRevMsg(msg)
			} else {
				log.Warn("r.newMsgRev[%s] !has", msg.Channel)
			}
		} else {
			return false, nil
		}
	case <-r.newChannel:
		log.Info("StartSubscription new channel rev")
	case <-time.After(30 * time.Minute):
		log.Info("StartSubscription <-sub.Channel() timeout for 30 min")
	}
	return true, nil
}
