package exec

import (
	"context"
	"sync"
)

// Parallelize will execute the provided function with the maximum
// specified level of parallelism until the context ends. Users of this
// function can use the passed sync.WaitGroup to wait for termination of
// all parallel tasks.
func Parallelize(ctx context.Context, wg *sync.WaitGroup, maxConcurrent int, f func()) {
	inProgressJobs := make(chan struct{}, maxConcurrent)
	for {
		select {
		case <-ctx.Done():
			close(inProgressJobs)
			return
		case inProgressJobs <- struct{}{}:
			wg.Add(1)
			go func() {
				defer wg.Done()
				f()
				<-inProgressJobs
			}()
		}
	}
}
