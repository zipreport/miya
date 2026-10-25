package runtime

import (
	"sync"
	"testing"
)

func TestEnterConcurrent(t *testing.T) {
	e := &DefaultEvaluator{}
	start := make(chan struct{})
	var wg sync.WaitGroup
	errs := make(chan error, 8*1000)
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for i := 0; i < 1000; i++ {
				leave, err := e.enter("t")
				if err != nil {
					errs <- err
				}
				leave()
			}
		}()
	}
	close(start)
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("unexpected error: %v", err)
	}
	if d := e.depth.Load(); d != 0 {
		t.Fatalf("depth = %d, want 0", d)
	}
}
