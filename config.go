package main

// nolint: lll
import (
	"github.com/brigadecore/badgr/internal/badges/redis"
	"github.com/brigadecore/brigade-foundations/http"
	"github.com/brigadecore/brigade-foundations/os"
)

// serverConfig populates configuration for the HTTP/S server from environment
// variables.
func serverConfig() (http.ServerConfig, error) {
	config := http.ServerConfig{}
	var err error
	config.Port, err = os.GetIntFromEnvVar("PORT", 8080)
	if err != nil {
		return config, err
	}
	config.TLSEnabled, err = os.GetBoolFromEnvVar("TLS_ENABLED", false)
	if err != nil {
		return config, err
	}
	if config.TLSEnabled {
		config.TLSCertPath, err = os.GetRequiredEnvVar("TLS_CERT_PATH")
		if err != nil {
			return config, err
		}
		config.TLSKeyPath, err = os.GetRequiredEnvVar("TLS_KEY_PATH")
		if err != nil {
			return config, err
		}
	}
	return config, nil
}

func redisCacheConfig() (redis.CacheConfig, error) {
	config := redis.CacheConfig{}
	var err error
	config.RedisHost, err = os.GetRequiredEnvVar("REDIS_HOST")
	if err != nil {
		return config, err
	}
	config.RedisPort, err = os.GetIntFromEnvVar("REDIS_PORT", 6379)
	if err != nil {
		return config, err
	}
	config.RedisPassword, err = os.GetRequiredEnvVar("REDIS_PASSWORD")
	if err != nil {
		return config, err
	}
	config.RedisDB, err = os.GetIntFromEnvVar("REDIS_DB", 0)
	if err != nil {
		return config, err
	}
	config.RedisEnableTLS, err = os.GetBoolFromEnvVar("REDIS_ENABLE_TLS", false)
	if err != nil {
		return config, err
	}
	config.RedisPrefix = os.GetEnvVar("REDIS_PREFIX", "")
	return config, nil
}
