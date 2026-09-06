package core

import (
	"sync"
	"testing"
)

// TestMemoryCacheConcurrent exercises the default in-process cache from many
// goroutines. Run under -race: an unguarded map would fail this test with a
// data race (regression guard for RF-001).
func TestMemoryCacheConcurrent(t *testing.T) {
	c := NewMemoryCache()
	ctx := t.Context()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := "k"
			if i%2 == 0 {
				key = "k2"
			}
			if err := c.SetWithTTL(ctx, key, "v", timeMinute); err != nil {
				t.Errorf("SetWithTTL: %v", err)
			}
			if _, ok, err := c.Get(ctx, key); err != nil || !ok {
				t.Errorf("Get: ok=%v err=%v", ok, err)
			}
		}(i)
	}
	wg.Wait()
}
