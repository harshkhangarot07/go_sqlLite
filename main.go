package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

type Movie struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Director string `json:"director"`
}

var db *sql.DB

func createMovie(w http.ResponseWriter, r *http.Request) {
	var movie Movie
	err := json.NewDecoder(r.Body).Decode(&movie)

	if err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}
	_, err = db.Exec("INSERT INTO movies (id, title, director) VALUES (?, ?, ?)", movie.ID, movie.Title, movie.Director)
	if err != nil {
		http.Error(w, "Failed to insert movie into database", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(movie)

}

func getMovies(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id , title , director FROM movies")
	if err != nil {
		http.Error(w, "failed to query database", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var movies []Movie

	for rows.Next() {
		var movie Movie
		if err := rows.Scan(&movie.ID, &movie.Title, &movie.Director); err != nil {
			http.Error(w, "failed to scan row", http.StatusInternalServerError)
			return
		}

		movies = append(movies, movie)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movies)
}

func getMovie(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	row := db.QueryRow("SELECT id, title, director FROM movies WHERE id = ?", params["id"])

	var movie Movie
	err := row.Scan(&movie.ID, &movie.Title, &movie.Director)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Movie not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to query database", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movie)
}

func main() {
	var err error
	db, err = sql.Open("sqlite3", "./movies.db")

	if err != nil {
		fmt.Println("failed to connect ot database:", err)
		return
	}

	defer db.Close()

	//create the movies table if it doesnt exist
	_, err = db.Exec(
		`CREATE TABLE IF NOT EXISTS movies(
			id TEXT PRIMARY KEY,
			title TEXT,
			director TEXT )
			`)

	if err != nil {
		fmt.Println("failed to create table:", err)
	}

	r := mux.NewRouter()

	r.HandleFunc("/movies", createMovie).Methods("POST")
	r.HandleFunc("/movies", getMovies).Methods("GET")
	r.HandleFunc("/movies/{id}", getMovie).Methods("GET")
	log.Fatal(http.ListenAndServe(":5060", r))

}
