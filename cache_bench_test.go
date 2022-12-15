package goche

import (
	"bytes"
	"testing"
)

func BenchmarkIntegration(b *testing.B) {
	data := blob('a', 1024)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cache := New[int, []byte]()

		for j := 0; j < 1000; j++ {
			cache.Set(j, data)
		}

		for j := 0; j < 1500; j++ {
			cache.Get(j)
		}
	}
}

func BenchmarkCache_Set(b *testing.B) {
	data := blob('a', 1024)

	cache := New[string, []byte]()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cache.Set("foo", data)
	}
}

func BenchmarkCache_Get(b *testing.B) {
	data := blob('a', 1024)

	cache := New[string, []byte]()
	cache.items["foo"] = cache.newItem(data)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cache.Get("foo")
	}
}

func BenchmarkCache_Delete(b *testing.B) {
	cache := New[string, []byte]()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cache.Delete("foo")
	}
}

func blob(char byte, len int) []byte {
	return bytes.Repeat([]byte{char}, len)
}
