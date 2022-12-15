package goche

import "time"

// cacheItem is struct that contains cached
// value and it's parameters.
type cacheItem[v value] struct {
	val      v
	cachedAt time.Time
	ttlReset bool
	ttl      time.Duration
}

// CacheItemOption is functional parameter
// that give ability to change cacheItem.
type CacheItemOption[v value] func(c *cacheItem[v])

// TTL return CacheItemOption with TTL (time to live).
// TTL parameter defines lifetime of cached value.
// When this time passes, the item is removed from the cache.
func TTL[v value](ttl time.Duration, reset ...bool) CacheItemOption[v] {
	return func(c *cacheItem[v]) {
		c.ttl = ttl

		if len(reset) > 0 {
			c.ttlReset = reset[0]
		}
	}
}

// TTLWithReset is the same as TTL, but TTL will
// reset after Cache.Get is called.
func TTLWithReset[v value](ttl time.Duration) CacheItemOption[v] {
	return TTL[v](ttl, true)
}

// newItem creates cacheItem and applies CacheItemOption(s).
func (c *Cache[k, v]) newItem(val v, opts ...CacheItemOption[v]) cacheItem[v] {
	item := cacheItem[v]{
		val:      val,
		cachedAt: c.timeNowFunc(),
	}

	for _, opt := range opts {
		opt(&item)
	}

	return item
}

// checkTTL checks TTL expire.
func (i *cacheItem[v]) checkTTL(time time.Time) bool {
	return i.ttl.Microseconds() == 0 || !time.After(i.cachedAt.Add(i.ttl))
}
