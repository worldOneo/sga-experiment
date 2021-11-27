package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/valyala/fasthttp"
)

func main() {
	client := &fasthttp.Client{}
	sql := benchmarkDB("sql", client, 100, 10)
	//mongo := benchmarkDB("mongo", client, 100, 10)
	scylla := benchmarkDB("scylla", client, 100, 10)

	logWriteStats("sql", sql)
	//logWriteStats("mongo", mongo)
	logWriteStats("Scylla", scylla)
}

type Stat struct {
	Errors       bool
	ResponseTime int64
}

type Stats []Stat

func benchmarkDB(dbtype string, client *fasthttp.Client, workerN, requestPerWorker int) Stats {
	endpoint := "http://127.0.0.1:8080/" + dbtype + "/todo/myTodo"
	body := `
	{
    "head": "Mein Todo",
    "desc": "Meine Beschreibung"
	}
	`
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(endpoint)
	req.Header.SetMethod(fasthttp.MethodPut)
	err := client.Do(req, nil)
	if err != nil {
		log.Fatalf("Setup todo-list for '%s': %v", dbtype, err)
	}

	stats := Stats{}
	for kRequest := 0; kRequest < 100; kRequest++ {
		wg := &sync.WaitGroup{}
		responses := make(chan Stat, 1_000)
		for workers := 0; workers < workerN; workers++ {
			wg.Add(1)
			go func(worker int) {
				for i := 0; i < requestPerWorker; i++ {
					req := fasthttp.AcquireRequest()
					res := fasthttp.AcquireResponse()

					req.SetRequestURI(endpoint)
					req.Header.SetMethod(fasthttp.MethodPost)
					req.Header.SetContentType("application/json")
					req.AppendBody([]byte(body))

					statErr := false
					before := time.Now()
					err := client.Do(req, res)
					if err != nil {
						statErr = true
						log.Printf("Err: %v", err)
					}
					responses <- Stat{statErr || res.StatusCode() != 200, time.Since(before).Nanoseconds()}
					fasthttp.ReleaseRequest(req)
					fasthttp.ReleaseResponse(res)
				}
				wg.Done()
			}(workers)
		}
		wg.Wait()
		close(responses)
		newStats := make(Stats, 1_000)
		i := 0
		for stat := range responses {
			newStats[i] = stat
			i++
		}
		stats = append(stats, newStats...)
	}
	return stats
}

func logWriteStats(dbname string, stats Stats) {
	if err := writeStats(dbname, stats); err != nil {
		log.Printf("Failed to write %s: %v", dbname, err)
	}
}

func writeStats(dbname string, stats Stats) error {
	file, err := os.Create(dbname + ".csv")
	if err != nil {
		return err
	}
	_, err = file.Write([]byte(fmt.Sprintf("%s requestN, %s Time [ns],%s Failed\n", dbname, dbname, dbname)))
	if err != nil {
		return err
	}
	for stat := range stats {
		_, err = file.Write([]byte(fmt.Sprintf("%d,%d,%t\n", stat, stats[stat].ResponseTime, stats[stat].Errors)))
		if err != nil {
			return err
		}
	}
	return nil
}
