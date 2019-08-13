package ccv3

import (
	"fmt"
	"sort"
	"time"
)

type ParallelJob struct {
	Id int
	Request requestOptions
}

type ParallelJobResult struct {
	Id int
}

type ParallelRequests struct {
	BatchSize int
	Jobs      []ParallelJob
}

func (client *Client) newParallelHTTPRequests(batchSize int, requestOptions []requestOptions) ParallelRequests {
	jobs := make([]ParallelJob, len(requestOptions))

	for i, request := range requestOptions {
		jobs[i] = ParallelJob {
			Id: i,
			Request: request,
		}
	}

	return ParallelRequests{
		BatchSize: batchSize,
		Jobs:      jobs,
	}
}

func (p ParallelRequests) MakeAll() {
	jobsChannel := make(chan ParallelJob, p.BatchSize)
	resultsChannel := make(chan ParallelJobResult, len(p.Jobs))
	results := make([]ParallelJobResult, len(p.Jobs))

	for i := 0; i < p.BatchSize; i++ {
		go worker(i, jobsChannel, resultsChannel)
	}

	for i := 0; i < len(p.Jobs); i++ {
		jobsChannel <- p.Jobs[i]
	}
	close(jobsChannel)

	for i := 0; i < len(p.Jobs); i++ {
		results[i] = <- resultsChannel
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Id < results[j].Id
	})
}

func worker(id int, jobsChannel chan ParallelJob, resultsChannel chan ParallelJobResult) {
	for job := range jobsChannel {
		fmt.Println("worker", id, "started  job", job.Id)

		time.Sleep(time.Second)

		fmt.Println("worker", id, "finished job", job.Id)

		resultsChannel <- ParallelJobResult{
			Id: job.Id,
		}
	}
}
