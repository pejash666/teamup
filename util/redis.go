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
	Logger.Println("redis init success")
}

func Redis() *redis.Client {
	return redisClient
}

func RedisSet(key string, value interface{}, exp time.Duration) error {
	_, err := Redis().Set(key, value, exp).Result()
	if err != nil {
		Logger.Printf("redis Set failed for key:%s, err:%v", key, err)
		return err
	}
	Logger.Printf("[RedisSet] success for key:%v", key)
	return nil
}

func RedisGet(key string) (string, error) {
	res, err := Redis().Get(key).Result()
	if err != nil && err != redis.Nil {
		Logger.Printf("redis Get failed for key:%s, err:%v", key, err)
		return "", err
	}
	Logger.Printf("[RedisGet] success for key:%v", key)
	return res, nil
}
