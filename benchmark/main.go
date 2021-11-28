package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	todo "github.com/worldOneo/database-demos"
)

func main() {
	sqlNvm, _ := todo.NewSQL()
	mongoNvm, _ := todo.NewMongo()
	scyllaNvm, _ := todo.NewScylla()

	sql := benchmarkDB(sqlNvm, 100, 10)
	logWriteStats("sql", sql)
	mongo := benchmarkDB(mongoNvm, 100, 10)
	logWriteStats("mongo", mongo)
	scylla := benchmarkDB(scyllaNvm, 100, 10)
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

func benchmarkDB(todonvm todo.TodoNvm, workerN, requestPerWorker int) Stats {

	todonvm.CreateList("myTodo")

	stats := Stats{
		Insert: []Stat{},
		Update: []Stat{},
		Delete: []Stat{},
	}

	log.Printf("Warming up")
	warmup := &sync.WaitGroup{}
	for kRequest := 0; kRequest < 1000; kRequest++ {
		warmup.Add(1)
		go func() {
			todo := todo.Todo{"", "Mein Todo", "Meine Beschreibung"}
			for i := 0; i < requestPerWorker; i++ {
				err := todonvm.Save("myTodo", &todo)
				if err != nil {
					log.Printf("Err: %v", err)
				}
			}
			warmup.Done()
		}()
	}
	warmup.Wait()
	log.Printf("Warmed up")

	ids := make([]string, 100*requestPerWorker*workerN)

	log.Printf("Inserting")
	// Insert
	for kRequest := 0; kRequest < 100; kRequest++ {
		wg := &sync.WaitGroup{}
		responses := make(chan Stat, requestPerWorker*workerN)
		for workers := 0; workers < workerN; workers++ {
			wg.Add(1)
			go func(worker int) {
				todo := todo.Todo{"", "Mein Todo", "Meine Beschreibung"}
				for i := 0; i < requestPerWorker; i++ {
					field := kRequest*worker*requestPerWorker + worker*requestPerWorker + i
					before := time.Now()
					err := todonvm.Save("myTodo", &todo)
					responses <- Stat{err != nil, time.Since(before).Nanoseconds()}
					ids[field] = todo.Id
					if err != nil {
						log.Printf("Err: %v", err)
					}
				}
				wg.Done()
			}(workers)
		}
		wg.Wait()
		log.Printf("Inserted")
		close(responses)
		newStats := make([]Stat, requestPerWorker*workerN)
		i := 0
		for stat := range responses {
			newStats[i] = stat
			i++
		}
		stats.Insert = append(stats.Insert, newStats...)
	}

	log.Printf("Writing")
	// Write
	for kRequest := 0; kRequest < 100; kRequest++ {
		wg := &sync.WaitGroup{}
		responses := make(chan Stat, requestPerWorker*workerN)
		for workers := 0; workers < workerN; workers++ {
			wg.Add(1)
			go func(worker int) {
				todo := todo.Todo{"", "Mein neues Todo", "Meine neue Beschreibung"}
				for i := 0; i < requestPerWorker; i++ {
					field := kRequest*worker*requestPerWorker + worker*requestPerWorker + i
					before := time.Now()
					todo.Id = ids[field]
					err := todonvm.Update("myTodo", todo)
					responses <- Stat{err != nil, time.Since(before).Nanoseconds()}
					if err != nil {
						log.Printf("Err: %v", err)
					}
				}
				wg.Done()
			}(workers)
		}
		wg.Wait()
		log.Printf("Written")
		close(responses)
		newStats := make([]Stat, requestPerWorker*workerN)
		i := 0
		for stat := range responses {
			newStats[i] = stat
			i++
		}
		stats.Update = append(stats.Update, newStats...)
	}

	log.Printf("Deleting")
	// Delete
	for kRequest := 0; kRequest < 100; kRequest++ {
		wg := &sync.WaitGroup{}
		responses := make(chan Stat, requestPerWorker*workerN)
		for workers := 0; workers < workerN; workers++ {
			wg.Add(1)
			go func(worker int) {
				todo := todo.Todo{"", "Mein neues Todo", "Meine neue Beschreibung"}
				for i := 0; i < requestPerWorker; i++ {
					field := kRequest*worker*requestPerWorker + worker*requestPerWorker + i
					before := time.Now()
					todo.Id = ids[field]
					err := todonvm.Delete("myTodo", todo)
					responses <- Stat{err != nil, time.Since(before).Nanoseconds()}
					if err != nil {
						log.Printf("Err: %v", err)
					}
				}
				wg.Done()
			}(workers)
		}
		wg.Wait()
		log.Printf("Deleted")
		close(responses)
		newStats := make([]Stat, requestPerWorker*workerN)
		i := 0
		for stat := range responses {
			newStats[i] = stat
			i++
		}
		stats.Delete = append(stats.Delete, newStats...)
	}
	log.Printf("Done")

	return stats
}

func logWriteStats(dbname string, stats Stats) {
	log.Printf("Writing %s.csv", dbname)
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
