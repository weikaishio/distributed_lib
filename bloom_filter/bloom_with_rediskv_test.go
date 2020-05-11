/*
 * Copyright (C) timwang  2020/5/11
 * Copyright (C) ixiaochuan.cn
 */
package bloom_filter

import (
	"fmt"
	"github.com/go-redis/redis"
	"strings"
	"testing"
	"time"
)
var (
	blr *BloomWithRedis
)

func init() {
	options := redis.Options{
		Addr:               "127.0.0.1:6379",
		Password:           "mypassword",
		DB:                 1,
		DialTimeout:        10 * time.Second,
		ReadTimeout:        10 * time.Second,
		WriteTimeout:       10 * time.Second,
		IdleTimeout:        60 * time.Second,
		IdleCheckFrequency: 15 * time.Second,
	}

	redisClient := redis.NewClient(&options)
	ping, err := redisClient.Ping().Result()
	if err != nil {
		_ = redisClient.Close()
		panic(fmt.Sprintf("Redis failed to ping, options:%v,err: %v", options, err))
	}
	if strings.ToLower(ping) != "pong" {
		panic(fmt.Sprintf("Redis unexpected ping response, pong:%s,options:%v", ping, options))
	}

	blr, err = NewBloomWithRedisWithMAndK(redisClient, 8000, 5, true)
	if err != nil {
		panic(fmt.Sprintf("NewBloomWithRedisWithMAndK err:%v\n", err))
	}
}
func TestBloomWithRedis_Contains(t *testing.T) {
	key := "testbf1"
	contains, err := blr.AddAndContains(key, uint64(1), uint64(2), uint64(3), uint64(5))
	t.Logf("contains:%v, err:%v", contains, err)
	contains, err = blr.Contains(key, uint64(1), uint64(3), uint64(6))
	t.Logf("contains 2:%v, err:%v", contains, err)

}

