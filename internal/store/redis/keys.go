package redis

import "fmt"

const (
	// KeyPrefixService is the prefix for service keys
	KeyPrefixService = "jump:service:"
	// KeyPrefixCache is the prefix for cache keys
	KeyPrefixCache = "jump:cache:"
	// KeyAllServices is the key for the set of all service IDs
	KeyAllServices = "jump:services:all"
)

// ServiceKey returns the Redis key for a service by ID
func ServiceKey(id string) string {
	return KeyPrefixService + id
}

// CacheKey returns the Redis key for a cached resolution
func CacheKey(query string) string {
	return KeyPrefixCache + query
}

// AllServicesKey returns the key for the set of all service IDs
func AllServicesKey() string {
	return KeyAllServices
}

// ExtractServiceID extracts the service ID from a Redis key
func ExtractServiceID(key string) (string, error) {
	if len(key) <= len(KeyPrefixService) {
		return "", fmt.Errorf("invalid service key: %s", key)
	}
	return key[len(KeyPrefixService):], nil
}
