package redis

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/go-redis/redis"
	"github.com/stretchr/testify/require"
)

func TestNewCache(t *testing.T) {
	const testPrefix = "foo"
	cache, ok := NewCache(
		CacheConfig{
			RedisPrefix:    testPrefix,
			RedisEnableTLS: true,
		},
	).(*cache)
	require.True(t, ok)
	require.Equal(t, testPrefix, cache.prefix)
	require.NotNil(t, cache.redisClient)
	require.NotNil(t, cache.getFn)
	require.NotNil(t, cache.setFn)
}

func TestSet(t *testing.T) {
	const testKey = "key"
	const testValue = "value"
	testCases := []struct {
		name       string
		cache      *cache
		assertions func(error)
	}{
		{
			name: "error writing to warm cache",
			cache: &cache{
				setFn: func(string, string, time.Duration) error {
					return errors.New("something went wrong")
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error writing result for key")
				require.Contains(t, err.Error(), "to warm cache")
			},
		},
		{
			name: "error writing to cold cache",
			cache: &cache{
				setFn: func(key string, _ string, _ time.Duration) error {
					if strings.Contains(key, "cold") {
						return errors.New("something went wrong")
					}
					return nil
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error writing result for key")
				require.Contains(t, err.Error(), "to cold cache")
			},
		},
		{
			name: "success",
			cache: &cache{
				setFn: func(string, string, time.Duration) error {
					return nil
				},
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.assertions(testCase.cache.Set(testKey, testValue))
		})
	}
}

func TestGet(t *testing.T) {
	const testKey = "key"
	const testValue = "value"
	testCases := []struct {
		name       string
		cache      *cache
		assertions func(string, error)
	}{
		{
			name: "error reading from cache",
			cache: &cache{
				getFn: func(key string) (string, error) {
					return "", errors.New("something went wrong")
				},
			},
			assertions: func(_ string, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error retrieving result for key")
				require.Contains(t, err.Error(), "from warm cache")
			},
		},
		{
			name: "cache miss",
			cache: &cache{
				getFn: func(key string) (string, error) {
					return "", redis.Nil
				},
			},
			assertions: func(value string, err error) {
				require.NoError(t, err)
				require.Empty(t, value)
			},
		},
		{
			name: "cache hit",
			cache: &cache{
				getFn: func(key string) (string, error) {
					return testValue, nil
				},
			},
			assertions: func(value string, err error) {
				require.NoError(t, err)
				require.Equal(t, testValue, value)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.assertions(testCase.cache.getInternal(testKey, true))
		})
	}
}
