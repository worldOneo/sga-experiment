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
	sql := benchmarkDB("sql", client)
	mongo := benchmarkDB("mongo", client)
	scylla := benchmarkDB("scylla", client)
	if err := writeStats("sql", sql); err != nil {
		log.Printf("Failed to write SQL:", err)
	}
	if err := writeStats("mongo", mongo); err != nil {
		log.Printf("Failed to write Mongo:", err)
	}
	if err := writeStats("Scylla", scylla); err != nil {
		log.Printf("Failed to write Scylla:", err)
	}
}

type Stat struct {
	Errors       bool
	ResponseTime int64
}

type Stats []Stat

func benchmarkDB(dbtype string, client *fasthttp.Client) Stats {
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
	for kRequest := 0; kRequest < 2; kRequest++ {
		wg := &sync.WaitGroup{}
		responses := make(chan Stat, 1_000)
		for workers := 0; workers < 100; workers++ {
			wg.Add(1)
			go func(worker int) {
				for i := 0; i < 10; i++ {
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
					log.Printf("%d done: %d, %s", (i+1)*worker, res.StatusCode(), string(res.Body()))
					responses <- Stat{statErr, time.Since(before).Nanoseconds()}
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
