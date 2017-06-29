package core

import (
	// Needed since we are using this for opening the connection
	"github.com/go-redis/redis"
)

type RedisState struct {
	redis      *redis.Client
	identifier string
}

func (rS *RedisState) New(redis *redis.Client, identifier string) {
	rS.redis = redis
	rS.identifier = identifier
}

func (rS *RedisState) Get(key string) string {
	stringCmd := rS.redis.HGet("redisState:"+rS.identifier, key)
	return stringCmd.Val()
}

func (rS *RedisState) HKeys() []string {
	stringSliceCmd := rS.redis.HKeys("redisState:" + rS.identifier)
	return stringSliceCmd.Val()
}

func (rS *RedisState) Set(key string, value string) error {
	statusCmd := rS.redis.HSet("redisState:"+rS.identifier, key, value)
	return statusCmd.Err()
}

func (rS *RedisState) SetM(set map[string]interface{}) error {
	statusCmd := rS.redis.HMSet("redisState:"+rS.identifier, set)
	return statusCmd.Err()
}

func (rS *RedisState) Delete() error {
	statusCmd := rS.redis.Del("redisState:" + rS.identifier)
	return statusCmd.Err()
}
