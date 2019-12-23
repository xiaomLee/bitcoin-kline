package common

import (
	"errors"

	"github.com/go-redis/redis"
)

type MapRedisCache struct {
	rediss map[string]*redis.Client
}

func NewMapRedisCache() *MapRedisCache {
	return &MapRedisCache{
		rediss: make(map[string]*redis.Client),
	}
}

func (c MapRedisCache) Close() {
	for _, r := range c.rediss {
		r.Close()
	}
}

// AddRedisInstance add a redis instance into poll
// note that, all instance should be added in init status of Application
func (c MapRedisCache) AddRedisInstance(key string, addr string, port string, pwd string, dbNum int) error {
	if _, ok := c.rediss[key]; !ok {
		redisdb := redis.NewClient(&redis.Options{
			Addr:       addr + ":" + port,
			Password:   pwd,
			DB:         dbNum,
			MaxRetries: 2, // retry 3 times (<=MaxRetries)
			PoolSize:   1024,
		})

		if _, err := redisdb.Ping().Result(); err == nil {
			c.rediss[key] = redisdb
		} else {
			return err
		}

	} else {
		return errors.New("repeated key")
	}

	return nil
}

func (c MapRedisCache) GetRedisInstance(key string) (*redis.Client, bool) {
	r, ok := c.rediss[key]
	return r, ok
}

// -----------------------------------------------------------------------------

var (
	defaultMapCache *MapRedisCache
)

func init() {
	defaultMapCache = NewMapRedisCache()
}

func AddRedisInstance(key string, addr string, port string, pwd string, dbNum int) error {
	return defaultMapCache.AddRedisInstance(key, addr, port, pwd, dbNum)
}

func GetRedisInstance(key string) (*redis.Client, bool) {
	return defaultMapCache.GetRedisInstance(key)
}

func MustGetRedisInstance(instance ...string) *redis.Client {
	name := ""
	if len(instance) == 1 {
		name = instance[0]
	}
	redisInstance, ok := GetRedisInstance(name)
	if !ok {
		panic("redis instance [" + name + "] not exists")
	}

	return redisInstance
}

func ReleaseRedisPool() {
	defaultMapCache.Close()
}
