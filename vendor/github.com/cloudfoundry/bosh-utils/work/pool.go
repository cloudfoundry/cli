package work

import (
	"sync"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type Pool struct {
	Count int
}

// ParallelDo Runs the given set of tasks in parallel using the configured number of worker go routines
// Will stop adding new tasks if a task throws an error, but will wait for in-flight tasks to finish
func (p Pool) ParallelDo(tasks ...func() error) error {
	jobs := make(chan func() error, len(tasks))
	errs := make(chan error, len(tasks))
	wg := &sync.WaitGroup{}

	wg.Add(p.Count)
	for i := 0; i < p.Count; i++ {
		p.spawnWorker(jobs, errs, wg)
	}

	for _, task := range tasks {
		jobs <- task
	}

	close(jobs)

	wg.Wait()

	close(errs)

	var combinedErrors []error
	for e := range errs {
		combinedErrors = append(combinedErrors, e)
	}

	if len(combinedErrors) > 0 {
		return bosherr.NewMultiError(combinedErrors...)
	}

	return nil
}

func (p Pool) spawnWorker(tasks <-chan func() error, errs chan<- error, wg *sync.WaitGroup) {
	go func() {
		for task := range tasks {
			err := task()
			if err != nil {
				errs <- err
				break
			}
		}

		wg.Done()
	}()
}
