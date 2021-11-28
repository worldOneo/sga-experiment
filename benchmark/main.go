package main

import (
	"bytes"
	"encoding/json"
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
	logWriteStats("sql", sql)
	mongo := benchmarkDB("mongo", client, 100, 10)
	logWriteStats("mongo", mongo)
	scylla := benchmarkDB("scylla", client, 100, 10)
	logWriteStats("Scylla", scylla)
}

type Stat struct {
	Errors       bool
	ResponseTime int64
}

type Stats struct {
	Insert []Stat
	Update []Stat
	Delete []Stat
}

type Todo struct {
	Id   string `json:"id,omitempty"`
	Head string `json:"head,omitempty"`
	Desc string `json:"desc,omitempty"`
}

func benchmarkDB(dbtype string, client *fasthttp.Client, workerN, requestPerWorker int) Stats {
	endpoint := "http://127.0.0.1:8080/" + dbtype + "/todo/myTodo"
	writeBody := `
	{
    "head": "Mein Todo",
    "desc": "Meine Beschreibung"
	}
	`
	updateBody := `
	{
		"id": "%s",
    "head": "Mein neues Todo",
    "desc": "Meine neue Beschreibung"
	}
	`

	req := fasthttp.AcquireRequest()
	req.SetRequestURI(endpoint)
	req.Header.SetMethod(fasthttp.MethodPut)
	err := client.Do(req, nil)
	if err != nil {
		log.Fatalf("Setup todo-list for '%s': %v", dbtype, err)
	}

	stats := Stats{
		Insert: []Stat{},
		Update: []Stat{},
		Delete: []Stat{},
	}

	ids := make([]string, 100*requestPerWorker*workerN)

	// Insert
	for kRequest := 0; kRequest < 100; kRequest++ {
		wg := &sync.WaitGroup{}
		responses := make(chan Stat, requestPerWorker*workerN)
		for workers := 0; workers < workerN; workers++ {
			wg.Add(1)
			go func(worker int) {
				var todo Todo
				for i := 0; i < requestPerWorker; i++ {
					field := kRequest*worker*requestPerWorker + worker*requestPerWorker + i
					req := fasthttp.AcquireRequest()
					res := fasthttp.AcquireResponse()

					req.SetRequestURI(endpoint)
					req.Header.SetMethod(fasthttp.MethodPost)
					req.Header.SetContentType("application/json")
					req.AppendBody([]byte(writeBody))

					statErr := false
					before := time.Now()
					err := client.Do(req, res)
					if err != nil {
						statErr = true
						log.Printf("Err: %v", err)
					}
					if res.StatusCode() == 200 {
						json.NewDecoder(bytes.NewReader(res.Body())).Decode(&todo)
						ids[field] = todo.Id
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
		newStats := make([]Stat, requestPerWorker*workerN)
		i := 0
		for stat := range responses {
			newStats[i] = stat
			i++
		}
		stats.Insert = append(stats.Insert, newStats...)
	}

	// Write
	for kRequest := 0; kRequest < 100; kRequest++ {
		wg := &sync.WaitGroup{}
		responses := make(chan Stat, requestPerWorker*workerN)
		for workers := 0; workers < workerN; workers++ {
			wg.Add(1)
			go func(worker int) {
				for i := 0; i < requestPerWorker; i++ {
					field := kRequest*worker*requestPerWorker + worker*requestPerWorker + i
					id := ids[field]
					req := fasthttp.AcquireRequest()
					res := fasthttp.AcquireResponse()

					req.SetRequestURI(endpoint)
					req.Header.SetMethod(fasthttp.MethodPatch)
					req.Header.SetContentType("application/json")
					req.AppendBody([]byte(fmt.Sprintf(updateBody, id)))

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
		newStats := make([]Stat, requestPerWorker*workerN)
		i := 0
		for stat := range responses {
			newStats[i] = stat
			i++
		}
		stats.Update = append(stats.Update, newStats...)
	}

	// Delete
	for kRequest := 0; kRequest < 100; kRequest++ {
		wg := &sync.WaitGroup{}
		responses := make(chan Stat, requestPerWorker*workerN)
		for workers := 0; workers < workerN; workers++ {
			wg.Add(1)
			go func(worker int) {
				for i := 0; i < requestPerWorker; i++ {
					field := kRequest*worker*requestPerWorker + worker*requestPerWorker + i
					id := ids[field]
					req := fasthttp.AcquireRequest()
					res := fasthttp.AcquireResponse()

					req.SetRequestURI(endpoint)
					req.Header.SetMethod(fasthttp.MethodDelete)
					req.Header.SetContentType("application/json")
					req.AppendBody([]byte(fmt.Sprintf(updateBody, id)))

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
		newStats := make([]Stat, requestPerWorker*workerN)
		i := 0
		for stat := range responses {
			newStats[i] = stat
			i++
		}
		stats.Delete = append(stats.Delete, newStats...)
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
	_, err = file.Write([]byte(fmt.Sprintf("%[1]s requestN,%[1]s insert Time [ns],%[1]s insert Failed,%[1]s update Time [ns],%[1]s update Failed,%[1]s delete Time [ns],%[1]s delete Failed\n", dbname)))
	if err != nil {
		return err
	}
	for stat := range stats.Insert {
		_, err = file.Write([]byte(fmt.Sprintf("%d,%d,%t,%d,%t,%d,%t\n", stat,
			stats.Insert[stat].ResponseTime, stats.Insert[stat].Errors,
			stats.Update[stat].ResponseTime, stats.Update[stat].Errors,
			stats.Delete[stat].ResponseTime, stats.Delete[stat].Errors,
		)))
		if err != nil {
			return err
		}
	}
	return nil
}
