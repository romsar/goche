package goche

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("time now func", func(t *testing.T) {
		cache := New[string, string]()

		want := time.Now().Unix()
		got := cache.timeNowFunc().Unix()

		if want != got {
			t.Fatalf("Got=%v, want=%v", got, want)
		}
	})
}

func TestCache_Set(t *testing.T) {
	t.Parallel()

	t.Run("basic test", func(t *testing.T) {
		t.Parallel()

		cache := New[string, string]()

		cache.Set("abc", "123")
		cache.Set("foo", "bar")

		gotSize := len(cache.items)
		wantSize := 2
		if gotSize != wantSize {
			t.Fatalf("Got size=%d, want size=%d", gotSize, wantSize)
		}

		gotVal := cache.items["foo"].val
		wantVal := "bar"
		if gotVal != wantVal {
			t.Fatalf("Got val=%s, want val=%s", gotVal, wantVal)
		}
	})

	t.Run("replace test", func(t *testing.T) {
		t.Parallel()

		cache := New[string, string]()

		cache.Set("foo", "zzz")
		cache.Set("foo", "bar")

		gotSize := len(cache.items)
		wantSize := 1
		if gotSize != wantSize {
			t.Fatalf("Got size=%d, want size=%d", gotSize, wantSize)
		}

		gotVal := cache.items["foo"].val
		wantVal := "bar"
		if gotVal != wantVal {
			t.Fatalf("Got val=%s, want val=%s", gotVal, wantVal)
		}
	})

	t.Run("set with ttl append to expires", func(t *testing.T) {
		t.Parallel()

		cache := New[string, string]()

		cache.Set("foo", "bar")
		gotSize := len(cache.expires)
		wantSize := 0
		if gotSize != wantSize {
			t.Fatalf("Got size=%d, want size=%d", gotSize, wantSize)
		}

		cache.Set("foo", "bar", TTL[string](1*time.Second))

		gotSize = len(cache.expires)
		wantSize = 1
		if gotSize != wantSize {
			t.Fatalf("Got size=%d, want size=%d", gotSize, wantSize)
		}

		_, ok := cache.expires["foo"]
		if !ok {
			t.Fatalf("Got ok=%t, want ok=%t", ok, !ok)
		}
	})

	t.Run("default ttl not applied when item has its own ttl", func(t *testing.T) {
		t.Parallel()

		cache := New[string, string](
			WithDefaultTTL[string, string](1 * time.Second),
		)

		cache.Set("foo", "bar", TTL[string](5*time.Second))

		gotTTL := cache.items["foo"].ttl.Seconds()
		wantTTL := float64(5)
		if gotTTL != wantTTL {
			t.Fatalf("Got seconds=%f, want seconds=%f", gotTTL, wantTTL)
		}
	})

	t.Run("default ttl apply", func(t *testing.T) {
		t.Parallel()

		cache := New[string, string](
			WithDefaultTTL[string, string](1 * time.Second),
		)

		cache.Set("foo", "bar")

		gotTTL := cache.items["foo"].ttl.Seconds()
		wantTTL := float64(1)
		if gotTTL != wantTTL {
			t.Fatalf("Got seconds=%f, want seconds=%f", gotTTL, wantTTL)
		}
	})

	t.Run("default ttl with reset", func(t *testing.T) {
		t.Parallel()

		cache := New[string, string](
			WithDefaultTTL[string, string](1*time.Second, true),
		)

		cache.Set("foo", "bar")

		gotTTLReset := cache.items["foo"].ttlReset
		if gotTTLReset != true {
			t.Fatalf("Got ttl reset=%t, want ttl reset=%t", gotTTLReset, !gotTTLReset)
		}
	})
}

func TestCache_Delete(t *testing.T) {
	t.Parallel()

	t.Run("basic test", func(t *testing.T) {
		t.Parallel()

		cache := New[string, string]()

		cache.items["foo"] = cache.newItem("bar")

		cache.Delete("foo")

		gotSize := len(cache.items)
		wantSize := 0
		if gotSize != wantSize {
			t.Fatalf("Got size=%d, want size=%d", gotSize, wantSize)
		}
	})

	t.Run("delete key not exists", func(t *testing.T) {
		t.Parallel()

		cache := New[string, string]()

		cache.items["foo"] = cache.newItem("bar")

		cache.Delete("abc")

		gotSize := len(cache.items)
		wantSize := 1
		if gotSize != wantSize {
			t.Fatalf("Got size=%d, want size=%d", gotSize, wantSize)
		}
	})

	t.Run("delete also delete key from expires", func(t *testing.T) {
		t.Parallel()

		cache := New[string, string]()

		cache.items["foo"] = cache.newItem("bar")
		cache.expires["foo"] = struct{}{}

		cache.Delete("foo")

		gotSize := len(cache.expires)
		wantSize := 0
		if gotSize != wantSize {
			t.Fatalf("Got size=%d, want size=%d", gotSize, wantSize)
		}
	})
}

func TestCache_Count(t *testing.T) {
	t.Parallel()

	t.Run("basic test", func(t *testing.T) {
		t.Parallel()

		cache := New[string, string]()

		gotSize := len(cache.items)
		wantSize := 0
		if gotSize != wantSize {
			t.Fatalf("Got size=%d, want size=%d", gotSize, wantSize)
		}

		cache.Set("foo", "bar")

		gotSize = len(cache.items)
		wantSize = 1
		if gotSize != wantSize {
			t.Fatalf("Got size=%d, want size=%d", gotSize, wantSize)
		}

		cache.Delete("foo")

		gotSize = len(cache.items)
		wantSize = 0
		if gotSize != wantSize {
			t.Fatalf("Got size=%d, want size=%d", gotSize, wantSize)
		}
	})
}

func TestCache_Get(t *testing.T) {
	t.Parallel()

	t.Run("get item that exists", func(t *testing.T) {
		t.Parallel()

		cache := New[string, string]()

		want := "bar"

		cache.items["foo"] = cache.newItem(want)
		got, ok := cache.Get("foo")
		if !ok {
			t.Fatalf("Got ok=%t, want ok=%t", ok, !ok)
		} else if got != want {
			t.Fatalf("Got=%s, want=%s", got, want)
		}
	})

	t.Run("get item that does not exists", func(t *testing.T) {
		t.Parallel()

		cache := New[string, string]()

		want := ""

		got, ok := cache.Get("foo")
		if ok {
			t.Fatalf("Got ok=%t, want ok=%t", ok, !ok)
		} else if got != want {
			t.Fatalf("Got=%s, want=%s", got, want)
		}
	})

	t.Run("get with ttl", func(t *testing.T) {
		t.Parallel()

		t.Run("ttl after or equal now, no reset", func(t *testing.T) {
			t.Parallel()

			cache := New[string, string]()
			cache.timeNowFunc = func() time.Time { return time.Unix(10, 0) }

			cache.items["foo"] = cache.newItem("bar", TTL[string](1*time.Second))

			cache.timeNowFunc = func() time.Time { return time.Unix(11, 0) }

			_, ok := cache.Get("foo")
			if !ok {
				t.Fatalf("Got ok=%t, want ok=%t", ok, true)
			}
		})

		t.Run("ttl before now, no reset", func(t *testing.T) {
			t.Parallel()

			cache := New[string, string]()
			cache.timeNowFunc = func() time.Time { return time.Unix(10, 0) }

			cache.items["foo"] = cache.newItem("bar", TTL[string](1*time.Second))

			cache.timeNowFunc = func() time.Time { return time.Unix(12, 0) }

			_, ok := cache.Get("foo")
			if ok {
				t.Fatalf("Got ok=%t, want ok=%t", ok, true)
			}
		})

		t.Run("ttl touch reset", func(t *testing.T) {
			t.Parallel()

			cache := New[string, string]()
			cache.timeNowFunc = func() time.Time { return time.Unix(10, 0) }

			cache.items["foo"] = cache.newItem("bar", TTLWithReset[string](100*time.Second))
			_, _ = cache.Get("foo")

			cache.timeNowFunc = func() time.Time { return time.Unix(12, 0) }
			_, _ = cache.Get("foo")

			got := cache.items["foo"].cachedAt.Unix()
			want := int64(12)

			if got != want {
				t.Fatalf("Got=%d, want=%d", got, want)
			}
		})

		t.Run("ttl not touch reset", func(t *testing.T) {
			t.Parallel()

			cache := New[string, string]()
			cache.timeNowFunc = func() time.Time { return time.Unix(10, 0) }

			cache.items["foo"] = cache.newItem("bar", TTL[string](100*time.Second))
			_, _ = cache.Get("foo")

			cache.timeNowFunc = func() time.Time { return time.Unix(12, 0) }
			_, _ = cache.Get("foo")

			got := cache.items["foo"].cachedAt.Unix()
			want := int64(10)

			if got != want {
				t.Fatalf("Got=%d, want=%d", got, want)
			}
		})
	})
}

func TestCache_Run(t *testing.T) {
	t.Parallel()

	t.Run("not deleting not expires", func(t *testing.T) {
		t.Parallel()

		cache := New[string, string](WithPollInterval[string, string](100 * time.Millisecond))

		cache.items["foo"] = cache.newItem("bar", TTL[string](1*time.Millisecond))

		ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
		defer cancel()

		cache.Run(ctx)

		gotSize := len(cache.items)
		wantSize := 1
		if gotSize != wantSize {
			t.Fatalf("Got size=%d, want size=%d", gotSize, wantSize)
		}
	})

	t.Run("not deleting expires", func(t *testing.T) {
		t.Parallel()

		cache := New[string, string](WithPollInterval[string, string](100 * time.Millisecond))

		cache.items["foo"] = cache.newItem("bar", TTL[string](1*time.Second))
		cache.expires["foo"] = struct{}{}

		ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
		defer cancel()

		cache.Run(ctx)

		gotItemsSize := len(cache.items)
		wantItemsSize := 1
		if gotItemsSize != wantItemsSize {
			t.Fatalf("Got items size=%d, want items size=%d", gotItemsSize, wantItemsSize)
		}

		gotExpiresSize := len(cache.expires)
		wantExpiresSize := 1
		if gotExpiresSize != wantExpiresSize {
			t.Fatalf("Got expires size=%d, want expires size=%d", gotItemsSize, wantItemsSize)
		}
	})

	t.Run("deleting expires", func(t *testing.T) {
		t.Parallel()

		cache := New[string, string](WithPollInterval[string, string](100 * time.Millisecond))

		cache.Set("foo", "bar", TTL[string](1*time.Millisecond))

		ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
		defer cancel()

		cache.Run(ctx)

		<-ctx.Done()

		gotItemsSize := cache.Count()
		wantItemsSize := 0
		if gotItemsSize != wantItemsSize {
			t.Fatalf("Got items size=%d, want items size=%d", gotItemsSize, wantItemsSize)
		}

		gotExpiresSize := cache.expiresCount()
		wantExpiresSize := 0
		if gotExpiresSize != wantExpiresSize {
			t.Fatalf("Got expires size=%d, want expires size=%d", gotItemsSize, wantItemsSize)
		}
	})
}

func Test_newCacheMapIfNil(t *testing.T) {
	t.Parallel()

	want := map[string]cacheItem[any]{}
	got := newCacheMapIfNil[string, any](want, 0)

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Got=%v, want=%v", got, want)
	}
}

func TestWithPollInterval(t *testing.T) {
	t.Parallel()

	want := 15 * time.Second

	cache := New[string, string](WithPollInterval[string, string](want))

	got := cache.pollInterval

	if want != got {
		t.Fatalf("Got=%v, want=%v", got, want)
	}
}

func TestWithValues(t *testing.T) {
	t.Parallel()

	want := map[string]string{"foo": "bar", "abc": "zzz"}

	cache := New[string, string](WithValues[string, string](want))

	got := cache.items

	if reflect.DeepEqual(got, want) {
		t.Fatalf("Got=%v, want=%v", got, want)
	}
}

func TestWithDefaultTTL(t *testing.T) {
	t.Parallel()

	want := 15 * time.Second

	cache := New[string, string](WithDefaultTTL[string, string](want))

	got := cache.defaultTTL

	if want != got {
		t.Fatalf("Got=%v, want=%v", got, want)
	}
}
