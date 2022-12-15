package goche

import (
	"context"
	"sync"
	"time"
)

// defaultPollInterval defines default poll interval.
//
// This affects the interval between the cycle in
// goroutine and clearing the cache of irrelevant elements.
const defaultPollInterval = 1 * time.Second

// key is type for key generic.
type key interface{ comparable }

// value is type for value generic.
type value interface{ any }

// Cache is struct that contain Cache and it's parameters.
type Cache[k key, v value] struct {
	mu              *sync.RWMutex
	items           map[k]cacheItem[v]
	timeNowFunc     func() time.Time
	expires         map[k]struct{}
	pollInterval    time.Duration
	defaultTTL      time.Duration
	defaultTTLReset bool
}

// Get returns item's value from Cache.
func (c *Cache[k, v]) Get(key k) (val v, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, ok := c.items[key]
	if !ok {
		return val, false
	}

	now := c.timeNowFunc()
	if !item.checkTTL(now) {
		return val, false
	}

	if item.ttlReset {
		item.cachedAt = now
		c.items[key] = item
	}

	return item.val, ok
}

// Set save value to Cache via creating cacheItem.
func (c *Cache[k, v]) Set(key k, val v, opts ...CacheItemOption[v]) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item := c.newItem(val, opts...)

	c.applyDefaultTTL(&item)
	c.addToExpires(key, item.ttl)

	c.items[key] = item
}

// Delete deletes Cache value by its key.
func (c *Cache[k, v]) Delete(key k) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.delete(key)
}

// Count returns length of Cache items.
func (c *Cache[k, v]) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.items)
}

// expiresCount returns count of expires.
func (c *Cache[k, v]) expiresCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.expires)
}

// delete is same as Delete, but without mutex locks.
func (c *Cache[k, v]) delete(key k) {
	delete(c.items, key)
	c.deleteFromExpires(key)
}

// Run start goroutine with workers.
// ATM there is workers to delete expires.
func (c *Cache[k, v]) Run(ctx context.Context) {
	interval := c.pollInterval
	if interval.Microseconds() == 0 {
		interval = defaultPollInterval
	}

	go c.pollWork(ctx, interval, c.deleteExpires)

	<-ctx.Done()
}

// deleteExpires delete expires and Cache items with expired TTL.
func (c *Cache[k, v]) deleteExpires() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for k := range c.expires {
		item, ok := c.items[k]
		if !ok {
			continue
		}

		if !item.checkTTL(c.timeNowFunc()) {
			c.delete(k)
		}
	}
}

// addToExpires add key to expires (if ttl does not contain zero-value).
func (c *Cache[k, v]) addToExpires(key k, ttl time.Duration) {
	if ttl.Microseconds() > 0 {
		c.expires[key] = struct{}{}
	}
}

// deleteFromExpires delete key from expires.
func (c *Cache[k, v]) deleteFromExpires(key k) {
	delete(c.expires, key)
}

// applyDefaultTTL applies default TTL value to cacheItem.
func (c *Cache[k, v]) applyDefaultTTL(item *cacheItem[v]) {
	if item.ttl.Microseconds() == 0 && c.defaultTTL.Microseconds() > 0 {
		item.ttl = c.defaultTTL
	}

	if !item.ttlReset && c.defaultTTLReset {
		item.ttlReset = true
	}
}

// pollWork do infinite poll work (until context get cancelled).
func (c *Cache[k, v]) pollWork(ctx context.Context, interval time.Duration, f func()) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			f()
		}
	}
}

// Option is functional option to modify Cache.
type Option[k key, v value] func(c *Cache[k, v])

// WithSize initialize map with capacity.
func WithSize[k key, v value](size int) Option[k, v] {
	return func(c *Cache[k, v]) {
		c.items = newCacheMapIfNil[k, v](c.items, size)
	}
}

// WithValues initialize map with values.
func WithValues[k key, v value](m map[k]v) Option[k, v] {
	return func(c *Cache[k, v]) {
		if len(m) == 0 {
			return
		}

		c.mu.Lock()
		defer c.mu.Unlock()

		c.items = make(map[k]cacheItem[v], len(m))
		for key, val := range m {
			c.items[key] = c.newItem(val)
		}
	}
}

// WithDefaultTTL initialize Cache with default TTL value.
// All cache items without TTL will have default TTL value, that was passed via this method.
func WithDefaultTTL[k key, v value](ttl time.Duration, reset ...bool) Option[k, v] {
	return func(c *Cache[k, v]) {
		c.defaultTTL = ttl

		if ttl.Microseconds() > 0 && len(reset) > 0 {
			c.defaultTTLReset = reset[0]
		}
	}
}

// WithPollInterval set pollInterval of Cache.
func WithPollInterval[k key, v value](interval time.Duration) Option[k, v] {
	return func(c *Cache[k, v]) {
		c.pollInterval = interval
	}
}

// New creates Cache and applies Option(s).
// Every time you call New - you need to specify types via generics.
// For example: Cache[int, string]{}
func New[k key, v value](opts ...Option[k, v]) *Cache[k, v] {
	c := &Cache[k, v]{
		mu:          &sync.RWMutex{},
		expires:     make(map[k]struct{}, 0),
		timeNowFunc: time.Now,
	}

	for _, opt := range opts {
		opt(c)
	}

	c.items = newCacheMapIfNil[k, v](c.items, 0)

	return c
}

// newCacheMapIfNil initializes Cache map with specified size.
func newCacheMapIfNil[k key, v value](m map[k]cacheItem[v], size int) map[k]cacheItem[v] {
	if len(m) > 0 {
		return m
	}

	return make(map[k]cacheItem[v], size)
}
