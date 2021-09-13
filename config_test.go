package main

// nolint: lll
import (
	"testing"

	"github.com/brigadecore/badgr/internal/badges/redis"
	"github.com/brigadecore/brigade-foundations/http"
	"github.com/stretchr/testify/require"
)

// Note that unit testing in Go does NOT clear environment variables between
// tests, which can sometimes be a pain, but it's fine here-- so each of these
// test functions uses a series of test cases that cumulatively build upon one
// another.

func TestServerConfig(t *testing.T) {
	testCases := []struct {
		name       string
		setup      func()
		assertions func(http.ServerConfig, error)
	}{
		{
			name: "PORT not an int",
			setup: func() {
				t.Setenv("PORT", "foo")
			},
			assertions: func(_ http.ServerConfig, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "was not parsable as an int")
				require.Contains(t, err.Error(), "PORT")
			},
		},
		{
			name: "TLS_ENABLED not a bool",
			setup: func() {
				t.Setenv("PORT", "8080")
				t.Setenv("TLS_ENABLED", "nope")
			},
			assertions: func(_ http.ServerConfig, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "was not parsable as a bool")
				require.Contains(t, err.Error(), "TLS_ENABLED")
			},
		},
		{
			name: "TLS_CERT_PATH required but not set",
			setup: func() {
				t.Setenv("TLS_ENABLED", "true")
			},
			assertions: func(_ http.ServerConfig, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "value not found for")
				require.Contains(t, err.Error(), "TLS_CERT_PATH")
			},
		},
		{
			name: "TLS_KEY_PATH required but not set",
			setup: func() {
				t.Setenv("TLS_CERT_PATH", "/var/ssl/cert")
			},
			assertions: func(_ http.ServerConfig, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "value not found for")
				require.Contains(t, err.Error(), "TLS_KEY_PATH")
			},
		},
		{
			name: "success",
			setup: func() {
				t.Setenv("TLS_KEY_PATH", "/var/ssl/key")
			},
			assertions: func(config http.ServerConfig, err error) {
				require.NoError(t, err)
				require.Equal(
					t,
					http.ServerConfig{
						Port:        8080,
						TLSEnabled:  true,
						TLSCertPath: "/var/ssl/cert",
						TLSKeyPath:  "/var/ssl/key",
					},
					config,
				)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup()
			config, err := serverConfig()
			testCase.assertions(config, err)
		})
	}
}

func TestRedisCacheConfig(t *testing.T) {
	testCases := []struct {
		name       string
		setup      func()
		assertions func(redis.CacheConfig, error)
	}{
		{
			name: "REDIS_HOST not set",
			assertions: func(_ redis.CacheConfig, err error) {
				require.Error(t, err)
				require.Contains(
					t,
					err.Error(),
					"value not found for required environment variable",
				)
				require.Contains(t, err.Error(), "REDIS_HOST")
			},
		},
		{
			name: "REDIS_PORT not an int",
			setup: func() {
				t.Setenv("REDIS_HOST", "localhost")
				t.Setenv("REDIS_PORT", "foo")
			},
			assertions: func(_ redis.CacheConfig, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "was not parsable as an int")
				require.Contains(t, err.Error(), "REDIS_PORT")
			},
		},
		{
			name: "REDIS_PASSWORD not set",
			setup: func() {
				t.Setenv("REDIS_PORT", "6379")
			},
			assertions: func(_ redis.CacheConfig, err error) {
				require.Error(t, err)
				require.Contains(
					t,
					err.Error(),
					"value not found for required environment variable",
				)
				require.Contains(t, err.Error(), "REDIS_PASSWORD")
			},
		},
		{
			name: "REDIS_DB not an int",
			setup: func() {
				t.Setenv("REDIS_PASSWORD", "foobar")
				t.Setenv("REDIS_DB", "foo")
			},
			assertions: func(_ redis.CacheConfig, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "was not parsable as an int")
				require.Contains(t, err.Error(), "REDIS_DB")
			},
		},
		{
			name: "REDIS_ENABLE_TLS not a bool",
			setup: func() {
				t.Setenv("REDIS_DB", "1")
				t.Setenv("REDIS_ENABLE_TLS", "foo")
			},
			assertions: func(_ redis.CacheConfig, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "was not parsable as a bool")
				require.Contains(t, err.Error(), "REDIS_ENABLE_TLS")
			},
		},
		{
			name: "success",
			setup: func() {
				t.Setenv("REDIS_ENABLE_TLS", "true")
				t.Setenv("REDIS_PREFIX", "foo")
			},
			assertions: func(config redis.CacheConfig, err error) {
				require.NoError(t, err)
				require.Equal(
					t,
					redis.CacheConfig{
						RedisHost:      "localhost",
						RedisPort:      6379,
						RedisPassword:  "foobar",
						RedisDB:        1,
						RedisEnableTLS: true,
						RedisPrefix:    "foo",
					},
					config,
				)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.setup != nil {
				testCase.setup()
			}
			config, err := redisCacheConfig()
			testCase.assertions(config, err)
		})
	}
}
