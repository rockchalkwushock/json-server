package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type (
	Todo struct {
		CreatedAt time.Time  `json:"created_at"`
		Done      bool       `json:"done"`
		ID        int        `json:"id"`
		Title     string     `json:"title"`
		UpdatedAt *time.Time `json:"updated_at"`
	}
)

func main() {
	var (
		mux    = http.NewServeMux()
		todos  = make(map[int]*Todo)
		todoMu sync.Mutex
	)

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "GET /\n")
	})

	mux.HandleFunc("DELETE /todos/{id}", func(w http.ResponseWriter, r *http.Request) {
		todoMu.Lock()
		defer todoMu.Unlock()

		id := r.PathValue("id")
		tid, err := strconv.Atoi(id)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}
		delete(todos, tid)
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /todos", func(w http.ResponseWriter, r *http.Request) {
		todoMu.Lock()
		defer todoMu.Unlock()

		ts := make([]*Todo, 0, len(todos))
		for _, t := range todos {
			ts = append(ts, t)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ts)
	})
	mux.HandleFunc("GET /todos/{id}", func(w http.ResponseWriter, r *http.Request) {
		todoMu.Lock()
		defer todoMu.Unlock()

		id := r.PathValue("id")
		tid, err := strconv.Atoi(id)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}
		t, ok := todos[tid]
		if !ok {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(t)
	})
	mux.HandleFunc("POST /todos", func(w http.ResponseWriter, r *http.Request) {
		var t Todo
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusInternalServerError)
			return
		}
		if err := json.Unmarshal(body, &t); err != nil {
			http.Error(w, "Error parsing request body", http.StatusBadRequest)
			return
		}

		todoMu.Lock()
		defer todoMu.Unlock()

		t.ID = len(todos) + 1
		t.CreatedAt = time.Now()
		t.Done = false
		t.UpdatedAt = &t.CreatedAt
		todos[t.ID] = &t
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(t)

	})
	mux.HandleFunc("PUT /todos/{id}", func(w http.ResponseWriter, r *http.Request) {
		var updates Todo
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusInternalServerError)
			return
		}
		if err := json.Unmarshal(body, &updates); err != nil {
			http.Error(w, "Error parsing request body", http.StatusBadRequest)
			return
		}

		todoMu.Lock()
		defer todoMu.Unlock()

		id := r.PathValue("id")
		tid, err := strconv.Atoi(id)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}
		t, ok := todos[tid]
		if !ok {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		now := time.Now()
		t.Done = updates.Done
		t.UpdatedAt = &now

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(t)
	})

	if err := http.ListenAndServe("localhost:8080", mux); err != nil {
		log.Fatalf("Server failed to open: %v", err)
	}
}
