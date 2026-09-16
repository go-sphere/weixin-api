package core

import (
	"sync"
	"testing"
	"time"
)

// TestMemoryCacheConcurrent exercises the default in-process cache from many
// goroutines. Run under -race: an unguarded map would fail this test with a
// data race (regression guard for RF-001).
func TestMemoryCacheConcurrent(t *testing.T) {
	c := NewMemoryCache()
	ctx := t.Context()
	var wg sync.WaitGroup
	for i := range 50 {
		wg.Go(func() {
			key := "k"
			if i%2 == 0 {
				key = "k2"
			}
			if err := c.SetWithTTL(ctx, key, "v", time.Minute); err != nil {
				t.Errorf("SetWithTTL: %v", err)
			}
			if _, ok, err := c.Get(ctx, key); err != nil || !ok {
				t.Errorf("Get: ok=%v err=%v", ok, err)
			}
		})
	}
	wg.Wait()
}
