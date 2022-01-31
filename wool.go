// wool is a Go library which implements a worker pool to run concurrent jobs.
package wool

import (
	"sync"
)

// Types and structs for jobs and results
type Job struct {
	Id     int
	Active bool
	Values ValueMap
}
type Result struct {
	Job    Job
	Values ValueMap
}
type ResultList []Result

// Value maps for jobs and results
type ValueMap map[string]string
type ValueMapList []map[string]string

// Channels for jobs and results
var jobChan = make(chan Job, 10)
var resChan = make(chan Result, 10)

// Functions to be implemented externally
var jobFunc func(Job) Result
var resFunc func(Result) Result

// Output result list
var jobResultList ResultList

// registerJobs reads the `jobValueMapList` variable and for each `valueMap`
// inside, creates a new `job` and sends it to channel `jobChan`. The `worker`
// goroutines will read this channel and process each job inside.
func registerJobs(jobValueMaps ValueMapList) {
	for i, valueMap := range jobValueMaps {
		job := Job{Id: i, Active: true, Values: valueMap}
		jobChan <- job
	}
	close(jobChan)
}

// resultProcessor is a wrapper to import and execute the externally implemented
// function `resFunc`. Function `Work` invokes `resulProcessor` as a goroutine
// which reads the channel `resChan` aiming to process and filter the results.
// If the result returned by `resFunc` has the field `Job.Active` set as false,
// it will be discarded. Otherwise, it will be appended to the `jobResultList`
// slice, which is returned by `Work` as output.
func resultProcessor(done chan bool) {
	for result := range resChan {
		result = resFunc(result)
		if result.Job.Active {
			jobResultList = append(jobResultList, result)
		}
	}
	done <- true
}

// worker is a wrapper to import and execute the externally implemented function
// `jobFunc`. Function `launchWorkerPool` invokes `worker` as a gorotuine which
// reads the channel `jobChan` aiming to proccess and filter the jobs. If the
// result returned by `jobFunc` has the field `Job.Active` set as `false`, it
// will be discarded. Otherwise, it will be sent to channel `resChan` to be
// processed by the `resultProcessor`.
func worker(wg *sync.WaitGroup) {
	for job := range jobChan {
		result := jobFunc(job)
		if result.Job.Active {
			resChan <- result
		}
	}
	wg.Done()
}

// launchWorkerPool spawns the worker pool, creating as many worker goroutines
// as required. `launchWorkerPool` is called by function `Work`.
func launchWorkerPool(workerNum int) {
	var wg sync.WaitGroup
	for i := 0; i < workerNum; i++ {
		wg.Add(1)
		go worker(&wg)
	}
	wg.Wait()
	close(resChan)
}

// Work function is the entrypoint to `wool`. As this function is exported, it
// can be called from outside the package. The arguments of this function are:
// - workerNum (int): The number of simultaneous workers to process the jobs.
// - jobValueMapList (Slice of maps): Value maps with data of each job.
// - jobFuncImpl (function): Externally implemented function to process jobs.
// - resFuncImpl (function): Externally implemented function to process results.
// Work returns `jobResultList` with the definitive results of the execution.
func Work(
	workerNum int,
	jobValueMapList []map[string]string,
	jobFuncImpl func(Job) Result,
	resFuncImpl func(Result) Result,
) []Result {
	jobFunc = jobFuncImpl
	resFunc = resFuncImpl
	done := make(chan bool)
	go registerJobs(jobValueMapList)
	go resultProcessor(done)
	launchWorkerPool(workerNum)
	<-done
	return jobResultList
}
