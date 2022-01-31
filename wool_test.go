package wool

import (
	"log"
	"strconv"
	"sync"
	"testing"
)

var (
	tables []struct {
		name   string
		class  string
		distAu float64
	}
	resultTable  map[string]string
	jobValueMaps ValueMapList
)

func validateJobResults() {
	for result := range resChan {
		result = resFunc(result)
		if result.Job.Active {
			planet := result.Job.Values["name"]
			if result.Values["distLy"] != resultTable[planet] {
				log.Fatalf("Error: Received distLy (%s) is different from the expected one (%s)\n",
					result.Values["distLy"], resultTable[planet])
			}
		}
	}
}

func TestWorker(t *testing.T) {
	// Clean shared variables
	jobChan = make(chan Job, 10)
	resChan = make(chan Result, 10)
	jobResultList = ResultList{}

	// Allocate new jobs
	go func() {
		for i, valueMap := range jobValueMaps {
			job := Job{Id: i, Active: true, Values: valueMap}
			jobChan <- job
		}
		close(jobChan)
	}()

	// Activate result processor
	done := make(chan bool)
	go func() {
		done <- true
	}()

	// Launch worker pool
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go worker(&wg)
	}
	wg.Wait()
	close(resChan)

	// Validate results
	validateJobResults()
}

func TestLaunchWorkerPool(t *testing.T) {
	// Clean shared variables
	jobChan = make(chan Job, 10)
	resChan = make(chan Result, 10)
	jobResultList = ResultList{}

	// Allocate new jobs
	go func() {
		for i, valueMap := range jobValueMaps {
			job := Job{Id: i, Active: true, Values: valueMap}
			jobChan <- job
		}
		close(jobChan)
	}()

	// Activate result processor
	done := make(chan bool)
	go func() {
		for result := range resChan {
			result = resFunc(result)
			if result.Job.Active {
				jobResultList = append(jobResultList, result)
			}
		}
		done <- true
	}()

	// Launch worker pool
	launchWorkerPool(10)

	// Validate results
	validateJobResults()
}

func TestRegisterJobs(t *testing.T) {
	// Clean shared variables
	jobChan = make(chan Job, 10)
	resChan = make(chan Result, 10)
	jobResultList = ResultList{}

	sentJobNum := len(jobValueMaps)
	gotJobNum := 0

	// Activate job processor
	done := make(chan bool)
	go func() {
		for range jobChan {
			gotJobNum = gotJobNum + 1
		}
		done <- true

		// Validate results
		if gotJobNum != sentJobNum {
			t.Errorf("Error: Expected (%d) results but got (%d)\n",
				sentJobNum, gotJobNum)
		}
	}()

	// Register jobs
	registerJobs(jobValueMaps)
}

func TestResultProcessor(t *testing.T) {
	// Clean shared variables
	jobChan = make(chan Job, 10)
	resChan = make(chan Result, 10)
	jobResultList = ResultList{}

	// Allocate dummy results
	go func() {
		for i := range jobValueMaps {
			result := Result{Job: Job{Id: i, Active: true, Values: ValueMap{}}, Values: ValueMap{}}
			resChan <- result
		}
		close(resChan)
	}()

	// Activate result processor
	done := make(chan bool)
	go resultProcessor(done)
	<-done

	// Validate results
	if len(jobResultList) != len(tables) {
		t.Errorf("Error: Expected (%d) results but got (%d)\n",
			len(tables), len(jobResultList))
	}
}

func TestWork(t *testing.T) {
	// Clean shared variables
	jobChan = make(chan Job, 10)
	resChan = make(chan Result, 10)
	jobResultList = ResultList{}

	// Launch worker pool
	Work(10, jobValueMaps, jobFunc, resFunc)

	// Validate results
	validateJobResults()
}

// init configures the environment to perform the testing.
// 1. Implements the external functions `jobFunc` and `resFunc` for `wool`
// 2. Adds sample data table to `tables` to be processed during the tests.
// 3. Adds a table with the expected result data to `resultTable.`
// 4. Generate `jobValueMaps` to send the values of all the jobs.
// data.
func init() {
	// Implement external function
	jobFunc = func(job Job) Result {
		values := map[string]string{}

		if job.Values["class"] == "planet" {
			distAu, _ := strconv.ParseFloat(job.Values["distAu"], 64)
			distLy := distAu / 63241
			values["distLy"] = strconv.FormatFloat(distLy, 'f', 6, 64)
		} else {
			job.Active = false
		}

		return Result{Job: job, Values: values}
	}

	// Implement external function
	resFunc = func(result Result) Result {
		// Get only inner solar system planets
		distAu, _ := strconv.ParseFloat(result.Job.Values["distAu"], 64)
		if distAu > 2 {
			result.Job.Active = false
		}
		return result
	}

	// Sample data table
	tables = []struct {
		name   string
		class  string
		distAu float64
	}{
		{"Mercury", "planet", 0.4},
		{"Venus", "planet", 0.7},
		{"Earth", "planet", 1},
		{"Mars", "planet", 1.5},
		{"Asteroid belt", "area", 2.3},
		{"Jupiter", "planet", 5.2},
		{"Saturn", "planet", 9.5},
		{"Uranus", "planet", 19.2},
		{"Neptune", "planet", 30.1},
		{"Kuiper belt", "area", 30},
		{"Heliosphere", "area", 123},
		{"Oort Cloud", "area", 50000},
	}

	// Expected result table
	resultTable = map[string]string{
		"Mercury": "0.000006",
		"Venus":   "0.000011",
		"Earth":   "0.000016",
		"Mars":    "0.000024",
	}

	// Generate jobValueMaps
	for _, table := range tables {
		valueMap := ValueMap{
			"name":   table.name,
			"class":  table.class,
			"distAu": strconv.FormatFloat(table.distAu, 'f', 2, 64),
		}
		jobValueMaps = append(jobValueMaps, valueMap)
	}
}
