package redis_pubsub

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/go-redis/redis"
	"github.com/mkideal/log"
)

var (
	redisClient *redis.Client
)

func init() {
	options := redis.Options{
		Addr:               "127.0.0.1:6379",
		Password:           "",
		DB:                 1,
		DialTimeout:        10 * time.Second,
		ReadTimeout:        10 * time.Second,
		WriteTimeout:       10 * time.Second,
		IdleTimeout:        60 * time.Second,
		IdleCheckFrequency: 15 * time.Second,
	}

	redisClient = redis.NewClient(&options)
	ping, err := redisClient.Ping().Result()
	if err != nil {
		_ = redisClient.Close()
		log.Fatal("Redis failed to ping, options:%v,err: %v", options, err)
	}
	if strings.ToLower(ping) != "pong" {
		log.Fatal("Redis unexpected ping response, pong:%s,options:%v", ping, options)
	}

}

type PubsubMsgTest struct {
	Title   string
	Content string
}

func (p PubsubMsgTest) MarshalBinary() (data []byte, err error) {
	return json.Marshal(p)
}
func TestRdsPubSubMsg_AddSubscribe(t *testing.T) {
	SharedRdsSubscribMsgInstance().AddSubscribe("test1", func(msg interface{}) {
		fmt.Printf("~~~~~~~~~~~~~~~exec:%v\n", msg)
	})
	SharedRdsSubscribMsgInstance().AddSubscribe("test2", func(msg interface{}) {
		fmt.Printf("~~~~~~~~~~~~~~~exec2:%v\n", msg)
	})
	go SharedRdsSubscribMsgInstance().StartSubscription()
	fmt.Println("TestRdsPubSubMsg_AddSubscribe")
	for {
		select {
		case <-time.After(3 * time.Second):
			_ = SharedRdsSubscribMsgInstance().Publish("test1", PubsubMsgTest{
				Title:   "test",
				Content: "content",
			})
			_ = SharedRdsSubscribMsgInstance().Publish("test2", PubsubMsgTest{
				Title:   "test2",
				Content: "content2",
			})
			SharedRdsSubscribMsgInstance().Set(redisClient)
			fmt.Println("publish")
			//if time.Now().Minute()%1 == 0 {
			//	SharedRdsSubscribMsgInstance().Quit()
			//	return
			//}
		}
	}
}
