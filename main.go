package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))
	r.Use(middleware.Timeout(10 * time.Second))
	sql, err := NewSQL()
	if err != nil {
		log.Fatalf("SQL %v", err)
	}
	mongo, err := NewMongo()
	if err != nil {
		log.Fatalf("Mongo %v", err)
	}
	addAdapter("/sql", sql, r)
	addAdapter("/mongo", mongo, r)
	http.ListenAndServe("localhost:8080", r)
}

// StatusMessage structure zum angeben des antwort status
// code http response code
// message grund
type StatusMessage struct {
	Status  int    `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
}

func addAdapter(name string, adapter TodoNvm, router chi.Router) {
	router.Route(name, func(r chi.Router) {
		r.Get("/todo", func(rw http.ResponseWriter, r *http.Request) {
			todos, err := adapter.Get()
			if err != nil {
				log.Printf("Get: %v", err)
				code(rw, http.StatusInternalServerError)
				return
			}
			rw.Header().Add("Content-Type", "application/json")
			rw.WriteHeader(http.StatusOK)
			write(rw, &todos)
		})

		r.Post("/todo", func(rw http.ResponseWriter, r *http.Request) {
			var todo Todo
			err := json.NewDecoder(r.Body).Decode(&todo)
			if err != nil {
				code(rw, http.StatusBadRequest)
				return
			}
			err = adapter.Save(&todo)
			if err != nil {
				log.Printf("Save: %v", err)
				code(rw, http.StatusInternalServerError)
				return
			}
			rw.Header().Add("Content-Type", "application/json")
			rw.WriteHeader(http.StatusOK)
			write(rw, &todo)
		})

		r.Delete("/todo", func(rw http.ResponseWriter, r *http.Request) {
			var todo Todo
			err := json.NewDecoder(r.Body).Decode(&todo)
			if err != nil {
				code(rw, http.StatusBadRequest)
				return
			}
			err = adapter.Delete(todo)
			if err != nil {
				log.Printf("Delete: %v", err)
				code(rw, http.StatusInternalServerError)
				return
			}
			rw.Header().Add("Content-Type", "application/json")
			rw.WriteHeader(http.StatusOK)
			write(rw, &todo)
		})

		r.Patch("/todo", func(rw http.ResponseWriter, r *http.Request) {
			var todo Todo
			err := json.NewDecoder(r.Body).Decode(&todo)
			if err != nil {
				code(rw, http.StatusBadRequest)
				return
			}
			err = adapter.Update(todo)
			if err != nil {
				log.Printf("Patch: %v", err)
				code(rw, http.StatusInternalServerError)
				return
			}
			rw.Header().Add("Content-Type", "application/json")
			rw.WriteHeader(http.StatusOK)
			write(rw, &todo)
		})
	})
}

func write(rw http.ResponseWriter, v interface{}) {
	json.NewEncoder(rw).Encode(v)
}

func code(rw http.ResponseWriter, statusCode int) {
	rw.WriteHeader(statusCode)
	write(rw, &StatusMessage{
		Status:  statusCode,
		Message: http.StatusText(statusCode),
	})
}
