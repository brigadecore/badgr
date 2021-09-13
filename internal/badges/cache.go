package badges

// Cache is the public interface for any component that can cache results.
type Cache interface {
	// Set writes a result to both warm and cold caches.
	Set(key, value string) error
	// Get reads a result from the warm cache. An empty string return value
	// indicates a cache miss.
	GetWarm(key string) (string, error)
	// Get reads a result from the cold cache. An empty string return value
	// indicates a cache miss.
	GetCold(key string) (string, error)
}
