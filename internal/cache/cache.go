package cache

import (
	"time"

	"github.com/go-redis/cache"
	"github.com/go-redis/redis"
	"github.com/tj/go/env"
	"github.com/vmihailenco/msgpack"
)

var (
	ring = redis.NewRing(&redis.RingOptions{
		Addrs: map[string]string{
			"saldotuc": env.Get("REDIS_HOST") + ":" + env.Get("REDIS_PORT"),
		},
		Password: env.Get("REDIS_PASSWORD"),
	})
	codec = &cache.Codec{
		Redis: ring,
		Marshal: func(v interface{}) ([]byte, error) {
			return msgpack.Marshal(v)
		},
		Unmarshal: func(b []byte, v interface{}) error {
			return msgpack.Unmarshal(b, v)
		},
	}
)

// Get gets the object for the given key.
func Get(key string, object interface{}) error {
	return codec.Get(key, object)
}

// Set caches the item.
func Set(key string, object interface{}) error {
	return codec.Set(&cache.Item{
		Expiration: time.Minute * 5,
		Key:        key,
		Object:     object,
	})
}
