package goche

import (
	"reflect"
	"testing"
	"time"
)

func TestCache_newItem(t *testing.T) {
	t.Parallel()

	t.Run("base test", func(t *testing.T) {
		t.Parallel()

		cache := New[string, string]()
		cache.timeNowFunc = func() time.Time { return time.Unix(5, 0) }

		got := cache.newItem("foo bar")
		want := cacheItem[string]{val: "foo bar", cachedAt: cache.timeNowFunc()}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("Got=%v, want=%v", got, want)
		}
	})

	t.Run("ttl test", func(t *testing.T) {
		t.Parallel()

		cache := New[string, string]()
		cache.timeNowFunc = func() time.Time { return time.Unix(0, 0) }

		got := cache.newItem("foo bar", TTL[string](1*time.Second))
		want := cacheItem[string]{val: "foo bar", cachedAt: cache.timeNowFunc(), ttl: 1 * time.Second}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("Got=%v, want=%v", got, want)
		}
	})

	t.Run("ttl with reset test", func(t *testing.T) {
		t.Parallel()

		cache := New[string, string]()
		cache.timeNowFunc = func() time.Time { return time.Unix(0, 0) }

		got := cache.newItem("foo bar", TTLWithReset[string](1*time.Second))
		want := cacheItem[string]{val: "foo bar", cachedAt: cache.timeNowFunc(), ttl: 1 * time.Second, ttlReset: true}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("Got=%v, want=%v", got, want)
		}
	})
}
