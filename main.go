package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
)

type Store struct {
	store map[string]string
	lock  sync.RWMutex
}

func GetData(path string) (string, string, error) {
	data := strings.Split(path, "=")
	if len(data) <= 1 {
		return "", "", fmt.Errorf("error missing data")
	}

	return data[0], data[1], nil
}

func Get(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, fmt.Sprintf("Invalid HTTP method received; expected GET, received %s", r.Method), 405)
			return
		}

		store.lock.RLock()
		defer store.lock.RUnlock()

		key := r.URL.Query().Get("key")
		if key == "" {
			http.Error(w, "Error missing key", 400)
			return
		}

		value, ok := store.store[key]
		if !ok {
			http.Error(w, fmt.Sprintf("Error missing key: %s", key), 404)
			return
		}

		w.Write([]byte(value))

		slog.Info("Retrieved", "key", key, "value", value)
	}
}

func Set(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, fmt.Sprintf("Invalid HTTP method received; expected POST, received %s", r.Method), 405)
			return
		}

		key, val, err := GetData(r.URL.RawQuery)
		if err != nil {
			slog.Info("error", "err", err)
			http.Error(w, "Error parsing key and value", 400)
			return
		}

		store.lock.Lock()
		defer store.lock.Unlock()

		slog.Info("Writing", "key", key, "value", val)
		store.store[key] = val

		w.WriteHeader(200)
	}
}

func NewServer(store *Store) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/get", Get(store))
	mux.HandleFunc("/set", Set(store))

	return mux
}

func main() {
	store := &Store{
		store: make(map[string]string),
	}

	if err := http.ListenAndServe(":8080", NewServer(store)); err != nil {
		panic(fmt.Errorf("error starting server: %w", err))
	}
}
