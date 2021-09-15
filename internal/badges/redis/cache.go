package redis

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/brigadecore/badgr/internal/badges"
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
)

const (
	tempCold = "cold"
	tempWarm = "warm"
)

// CacheConfig represents configuration options for the Redis-based
// implementation of the badges.Cache interface
type CacheConfig struct {
	RedisHost      string
	RedisPort      int
	RedisPassword  string
	RedisDB        int
	RedisEnableTLS bool
	RedisPrefix    string
}

type cache struct {
	redisClient *redis.Client
	prefix      string
	// The following internal functions are overridable for testing purposes
	getFn func(key string) (string, error)
	setFn func(key, value string, ttl time.Duration) error
}

// NewCache returns a new Redis-based implementation of the badges.Cache
// interface.
func NewCache(config CacheConfig) badges.Cache {
	redisOpts := &redis.Options{
		Addr:       fmt.Sprintf("%s:%d", config.RedisHost, config.RedisPort),
		Password:   config.RedisPassword,
		DB:         config.RedisDB,
		MaxRetries: 5,
	}
	if config.RedisEnableTLS {
		redisOpts.TLSConfig = &tls.Config{
			ServerName: config.RedisHost,
		}
	}
	cache := &cache{
		redisClient: redis.NewClient(redisOpts),
		prefix:      config.RedisPrefix,
	}
	cache.getFn = cache.get
	cache.setFn = cache.set
	return cache
}

func (c *cache) Set(key, value string) error {
	warmKey := c.getKey(key, true)
	if err := c.setFn(warmKey, value, time.Minute); err != nil {
		return errors.Wrapf(
			err,
			"error writing result for %s to warm cache",
			key,
		)
	}
	coldKey := c.getKey(key, false)
	if err := c.setFn(coldKey, value, 24*time.Hour); err != nil {
		return errors.Wrapf(
			err,
			"error writing result for key %q to cold cache",
			key,
		)
	}
	return nil
}

func (c *cache) GetWarm(key string) (string, error) {
	return c.getInternal(key, true)
}

func (c *cache) GetCold(key string) (string, error) {
	return c.getInternal(key, false)
}

func (c *cache) getInternal(key string, warm bool) (string, error) {
	temp := tempCold
	if warm {
		temp = tempWarm
	}
	key = c.getKey(key, warm)
	value, err := c.getFn(key)
	if err == redis.Nil {
		return "", nil // This isn't an error; it's just a cache miss
	} else if err != nil {
		return "", errors.Wrapf(
			err,
			"error retrieving result for key %q from %s cache",
			key,
			temp,
		)
	}
	return value, nil
}

func (c *cache) getKey(key string, warm bool) string {
	temp := tempCold
	if warm {
		temp = tempWarm
	}
	key = fmt.Sprintf("%s:%s", temp, key)
	if c.prefix == "" {
		return key
	}
	return fmt.Sprintf("%s:%s", c.prefix, key)
}

func (c *cache) get(key string) (string, error) {
	strCmd := c.redisClient.Get(key)
	return strCmd.Val(), strCmd.Err()
}

func (c *cache) set(key, value string, ttl time.Duration) error {
	return c.redisClient.Set(key, value, ttl).Err()
}
