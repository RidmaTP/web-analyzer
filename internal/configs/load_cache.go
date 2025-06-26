package configs
import (
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
)
//cache configuration using go-cache
//used to temporarily save api responses in cache for lightning fast responses.

// cache once ensures idempotency
// threadsafe
var (
	cacheOnce  sync.Once
	cacheVar *cache.Cache
)

// cache will be cleaned up every 2 hr with a 10 min cleanup interval
func LoadCacheConfig() *cache.Cache {
	cacheOnce.Do(func() {
		cacheVar = cache.New(2*time.Hour, 10*time.Minute)
	})
	return cacheVar
}

func GetCacheConfig() *cache.Cache{
	if cacheVar == nil {
		LoadCacheConfig()
	}
	return cacheVar
}