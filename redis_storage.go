package main

import "gopkg.in/redis.v3"

type redisStorage struct {
	client *redis.Client
	prefix string
}

type redisConfiguration struct {
	redis.Options
	Prefix string
}

func (r *redisStorage) Open(c *shortieConfiguration) (err error) {
	redisConfig := redisConfiguration{
		Prefix: "shortie:",
		Options: redis.Options{
			Addr: "localhost:6379",
		},
	}

	if err = c.UnifyStorageConfiguration("redis", &redisConfig); err != nil {
		return
	}

	r.prefix = redisConfig.Prefix
	r.client = redis.NewClient(&redisConfig.Options)

	_, err = r.client.Ping().Result()
	return
}

func (r *redisStorage) Close() error {
	return r.client.Close()
}

func (r *redisStorage) Store(key, value string) error {
	return r.client.Set(r.prefix+key, value, 0).Err()
}

func (r *redisStorage) Fetch(key string) (string, error) {
	result := r.client.Get(r.prefix + key)
	return result.String(), result.Err()
}

func init() {
	RegisterStorageInterface("redis", func() StorageInterface {
		return &redisStorage{}
	})
}
