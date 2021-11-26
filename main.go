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
		r.Get("/todo/{list}", func(rw http.ResponseWriter, r *http.Request) {
			list := chi.URLParam(r, "list")
			todos, err := adapter.Get(list)
			if err != nil {
				log.Printf("Get: %v", err)
				code(rw, http.StatusInternalServerError)
				return
			}
			write(rw, &todos, http.StatusOK)
		})

		r.Post("/todo/{list}", func(rw http.ResponseWriter, r *http.Request) {
			list := chi.URLParam(r, "list")
			var todo Todo
			err := json.NewDecoder(r.Body).Decode(&todo)
			if err != nil {
				code(rw, http.StatusBadRequest)
				return
			}
			err = adapter.Save(list, &todo)
			if err != nil {
				log.Printf("Save: %v", err)
				code(rw, http.StatusInternalServerError)
				return
			}
			write(rw, &todo, http.StatusOK)
		})

		r.Delete("/todo/{list}", func(rw http.ResponseWriter, r *http.Request) {
			list := chi.URLParam(r, "list")
			var todo Todo
			err := json.NewDecoder(r.Body).Decode(&todo)
			if err != nil {
				code(rw, http.StatusBadRequest)
				return
			}
			err = adapter.Delete(list, todo)
			if err != nil {
				log.Printf("Delete: %v", err)
				code(rw, http.StatusInternalServerError)
				return
			}
			write(rw, &todo, http.StatusOK)
		})

		r.Patch("/todo/{list}", func(rw http.ResponseWriter, r *http.Request) {
			list := chi.URLParam(r, "list")
			var todo Todo
			err := json.NewDecoder(r.Body).Decode(&todo)
			if err != nil {
				code(rw, http.StatusBadRequest)
				return
			}
			err = adapter.Update(list, todo)
			if err != nil {
				log.Printf("Patch: %v", err)
				code(rw, http.StatusInternalServerError)
				return
			}
			write(rw, &todo, http.StatusOK)
		})

		r.Put("/todo/{list}", func(rw http.ResponseWriter, r *http.Request) {
			list := chi.URLParam(r, "list")
			err := adapter.CreateList(list)
			if err != nil {
				log.Printf("Put: %v", err)
				code(rw, http.StatusInternalServerError)
				return
			}
			code(rw, http.StatusOK)
		})

		r.Options("/todo/{list}/{name}", func(rw http.ResponseWriter, r *http.Request) {
			list := chi.URLParam(r, "list")
			name := chi.URLParam(r, "name")
			err := adapter.RenameList(list, name)

			if err != nil {
				log.Printf("Options: %v", err)
				code(rw, http.StatusInternalServerError)
				return
			}
			code(rw, http.StatusOK)
		})
	})
}

func write(rw http.ResponseWriter, v interface{}, code int) {
	rw.Header().Add("Content-Type", "application/json")
	rw.WriteHeader(code)
	json.NewEncoder(rw).Encode(v)
}

func code(rw http.ResponseWriter, statusCode int) {
	write(rw, &StatusMessage{
		Status:  statusCode,
		Message: http.StatusText(statusCode),
	}, statusCode)
}
