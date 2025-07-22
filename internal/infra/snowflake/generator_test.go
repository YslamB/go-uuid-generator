package snowflake

import (
	"sync"
	"testing"
)

func BenchmarkSnowflakeGenerator(b *testing.B) {
	gen, _ := NewGenerator(1, 1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gen.NextID()
	}
}

func BenchmarkSnowflakeParallel(b *testing.B) {
	gen, _ := NewGenerator(1, 1)
	var mu sync.Mutex
	ids := make(map[int64]struct{})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			id, _ := gen.NextID()
			mu.Lock()
			// Check for duplicates
			if _, exists := ids[id]; exists {
				b.Errorf("Duplicate ID found: %d", id)
			}
			ids[id] = struct{}{}
			mu.Unlock()
		}
	})
}
