package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

// HTTP Handlers

func HandleIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, info())
}

func HandleStart(w http.ResponseWriter, r *http.Request) {
	state, err := decodeState(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	start(state)
	w.WriteHeader(http.StatusOK)
}

func HandleMove(w http.ResponseWriter, r *http.Request) {
	state, err := decodeState(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, move(state))
}

func HandleEnd(w http.ResponseWriter, r *http.Request) {
	state, err := decodeState(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	end(state)
	w.WriteHeader(http.StatusOK)
}

func decodeState(r *http.Request) (GameState, error) {
	var state GameState
	err := json.NewDecoder(r.Body).Decode(&state)
	return state, err
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("ERROR: Failed to encode response, %s", err)
	}
}

// Middleware

// Method matching lives in the handlers rather than the route patterns because
// pattern-based matching silently degrades to path-literal matching on Go before 1.22.
func requirePost(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		next(w, r)
	}
}

// Start Battlesnake Server

func newMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", HandleIndex)
	mux.HandleFunc("/start", requirePost(HandleStart))
	mux.HandleFunc("/move", requirePost(HandleMove))
	mux.HandleFunc("/end", requirePost(HandleEnd))
	return mux
}

func RunServer() {
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8000"
	}

	log.Printf("Running Battlesnake at http://0.0.0.0:%s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, newMux()))
}
