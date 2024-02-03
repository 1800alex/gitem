package main

import (
	"fmt"
	"time"

	"gitm/workerpool"
)

func main() {
	numWorkers := 2
	numJobs := 10

	// Create a worker pool with failFast option.
	// wp := workerpool.New[int](numWorkers, true)
	// wp.Start()
	wp := workerpool.Go[int](numWorkers, false)

	go func() {
		// Add jobs to the worker pool.
		for i := 0; i < numJobs; i++ {
			jobNum := i
			fmt.Println("Adding job", i)
			wp.AddJob(func() (int, error) {
				if jobNum == 5 {
					return jobNum, fmt.Errorf("Job %d failed", jobNum)
				}
				fmt.Printf("Job %d started\n", jobNum)
				// Simulate some work
				time.Sleep(time.Millisecond * 100)
				fmt.Printf("Job %d completed\n", jobNum)
				return jobNum, nil
			})
		}

		wp.Done()
		fmt.Println("All jobs completed")
	}()

	// Process results
	for result := range wp.Results() {
		if result.Err != nil {
			fmt.Printf("Job failed with error: %v\n", result.Err)
		} else {
			fmt.Printf("Job completed with result: %d\n", result.Result)
		}
	}
	wp.Wait()

	fmt.Println("Done")
	time.Sleep(time.Second * 1)
}
