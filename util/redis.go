package util

import (
	"github.com/go-redis/redis"
	"sync"
	"time"
)

var (
	redisClient *redis.Client
	onceRedis   sync.Once
)

func InitRedis() {
	if redisClient != nil {
		return
	}
	onceRedis.Do(func() {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		})
	})
}

func Redis() *redis.Client {
	return redisClient
}

func RedisSet(key string, value interface{}, exp time.Duration) error {
	_, err := redisClient.Set(key, value, exp).Result()
	if err != nil {
		Logger.Printf("redis Set failed for key:%s, err:%v", key, err)
		return err
	}
	return nil
}

func RedisGet(key string) (string, error) {
	res, err := redisClient.Get(key).Result()
	if err != nil && err != redis.Nil {
		Logger.Printf("redis Get failed for key:%s, err:%v", key, err)
		return "", err
	}
	return res, nil
}
