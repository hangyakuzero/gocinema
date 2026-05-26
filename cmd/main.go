package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.Handle("GET /", http.FileServer(http.Dir("static")))
	mux.HandleFunc("GET /movies", listMovies)

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}

var movies = []moviesResponse{
	{ID: "dune", Title: "Dune", Rows: 6, SeatsPerRow: 10},
	{ID: "odyssey", Title: "The Odyssey", Rows: 7, SeatsPerRow: 12},
	{ID: "interstellar", Title: "Interstellar", Rows: 8, SeatsPerRow: 10},
	{ID: "blade-runner-2049", Title: "Blade Runner 2049", Rows: 6, SeatsPerRow: 9},
	{ID: "arrival", Title: "Arrival", Rows: 5, SeatsPerRow: 8},
	{ID: "oppenheimer", Title: "Oppenheimer", Rows: 7, SeatsPerRow: 11},
	{ID: "tenet", Title: "Tenet", Rows: 6, SeatsPerRow: 10},
	{ID: "mad-max-fury-road", Title: "Mad Max: Fury Road", Rows: 5, SeatsPerRow: 9},
	{ID: "the-batman", Title: "The Batman", Rows: 6, SeatsPerRow: 10},
	{ID: "inception", Title: "inception", Rows: 5, SeatsPerRow: 8},
}

func listMovies(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, movies)
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

type moviesResponse struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Rows        int    `json:"rows"`
	SeatsPerRow int    `json:"seats_per_row"`
}
