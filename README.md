# wool

Worker Pool implementation for Goroutines, the name comes from the fusion of the words **Wo**rker and Po**ol**. This implementation is based in the code released by [Naveen Ramanathan](https://golangbot.com/about/) in the [Buffered Channels and Worker Pools](https://golangbot.com/buffered-channels-worker-pools/) article. I liked the snipped and I thought it was an excellent base to convert it into a module. This project may have sporadic updates as I need new features.

## Installation

Install wool:

```
$ go get github.com/otzarri/wool
```

## Usage

Import wool into your program and use it. The example below shows how to create a very simple concurrent application. The `jobFunc` and `resfunc` functions must be externally implemented to allow `wool` to execute custom code.

```
$ mkdir app
$ cd app
$ vim app.go
```

```go
package main

import (
        "fmt"
        "strconv"

        "github.com/otzarri/wool"
)

// jobFunc is a function implementation expected by `wool`. When a new job is
// allocated, `wool` sends the job to `jobFunc`, where it is expected to be an
// implementation to process the job and return a result. The result will be
// received by `wool` and if `Job.Active` field is `false`, it will be discarded.
// Otherwise, the result will be directed to the function `resFunc`.
func jobFunc(job wool.Job) wool.Result {
        values := map[string]string{}

        number, err := strconv.Atoi(job.Values["number"])
        if err != nil {
                job.Active = false
        }
        if number%2 == 0 {
                values["type"] = "even"
        } else {
                values["type"] = "odd"
        }

        return wool.Result{Job: job, Values: values}
}

// resFunc is a function implementation expected by `wool`. When a new result is
// returned from `jobFunc`, `wool` sends the result to `resFunc`, where it is
// expected to be an implementation to process the result and return a it. The
// result will be received by `wool` and if `Result.Job.Active` field is
// `false`, it will be discarded. Otherwise, the result will be added to the
// result list returned by the function `wool.Work`, executed into `main()`.
func resFunc(result wool.Result) wool.Result {
        if result.Values["type"] == "odd" {
                result.Job.Active = false
        }
        return result
}

func main() {
        workers := 10
        minNum := 0
        maxNum := 99

        jobValueMapList := []map[string]string{}

        for i := minNum; i < maxNum; i++ {
                jobValueMapList = append(jobValueMapList, map[string]string{"number": strconv.Itoa(i)})
        }

        jobResultList := wool.Work(workers, jobValueMapList, jobFunc, resFunc)
        fmt.Printf("The even numbers between %d and %d are: ", minNum, maxNum)
        for _, result := range jobResultList {
                number := result.Job.Values["number"]
                fmt.Printf("%s ", number)
        }
        fmt.Println()
}
```

```
$ go mod init app
$ go mod tidy
$ go run app.go
The even numbers between 0 and 99 are: 0 2 4 6 8 10 20 22 24 26 12 14 18 28 30 32 34 16 36 38 40 42 44 46 48 50 52 54 56 58 60 62 74 76 78 80 82 84 64 66 68 70 72 86 88 90 92 94 96 9
```

⭐ If you want to see `wool` working into a program, you can take a look at the [otzarri/nscan](http://github.com/otzarri/nscan) project.

## Testing

Run tests as usual in Go:

```
$ go test -v ./...
```

Run coverage tests:

```
$ go test -v -coverprofile=coverage.out ./...
$ go tool cover -html=coverage.out
```
