package worker

import (
	"context"
	"sync"
	"time"
)

// StartWorkerPool initializes the workers.
// NOW accepts: 'client' (to delete) and 'isDryRun' (to know if we should delete).
func StartWorkerPool(client S3Deleter, numWorkers int, daysOld int, isDryRun bool, jobs <-chan Job) <-chan Result {
	results := make(chan Result, numWorkers)
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				// Pass the client and dryRun flag to process
				results <- process(client, job, daysOld, isDryRun)
			}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	return results
}

func process(client S3Deleter, job Job, daysOld int, isDryRun bool) Result {
	age := time.Since(job.LastModified)
	threshold := time.Duration(daysOld) * 24 * time.Hour
	isStale := age > threshold

	var err error

	// THE DESTRUCTION LOGIC
	if isStale && !isDryRun {
		// Actually Delete!
		err = client.DeleteObject(context.TODO(), job.Bucket, job.Key)
	}

	return Result{
		Key:     job.Key,
		IsStale: isStale,
		Size:    job.Size,
		Err:     err,
	}
}
