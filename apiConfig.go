package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
)

type apiConfig struct {
	fileserverHits int
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits += 1
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Hits: %v", cfg.fileserverHits)
}

func chirpsPost(w http.ResponseWriter, r *http.Request, db *DB) {
	type requestBodyParams struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	bodyFetched := requestBodyParams{}
	err := decoder.Decode(&bodyFetched)

	// now the content of the request body is in bodyFetched variable !

	if err != nil {
		http.Error(w, "Something went wrong!", http.StatusBadRequest)
	}

	chirp := bodyFetched.Body
	// the chirp must be less than 140
	if len(chirp) > 140 {
		http.Error(w, "Chirp is too long !", http.StatusBadRequest)
	}
	// Remove dirty words from the chirp
	dirtyWords := map[string]bool{
		"kerfuffle": true,
		"sharbert":  true,
		"fornax":    true,
	}
	words := strings.Fields(chirp)
	var cleanedChirp strings.Builder
	for _, word := range words {
		lowerWord := strings.ToLower(word)
		if dirtyWords[lowerWord] {
			cleanedChirp.WriteString(strings.Repeat("*", 4))
		} else {
			cleanedChirp.WriteString(word)
		}
		cleanedChirp.WriteString(" ")
	}

	// Now our chirp is valid, its time to create the response !
	// Chirp is a strucutre !
	cleanedChirpStr := strings.TrimSpace(cleanedChirp.String())

	// Now our chirp is valid, it's time to create the response!
	responseBody, err := db.CreateChirp(cleanedChirpStr)
	if err != nil {
		return
	}

	// Marshal the response into JSON
	responseJSON, err := json.Marshal(responseBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Set the response headers and write the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(responseJSON)

}

func chirpsGet(w http.ResponseWriter, r *http.Request, db *DB) {
	chirps, err := db.GetChirps()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Sort chirps by id in ascending order
	sort.Slice(chirps, func(i, j int) bool {
		return chirps[i].Id < chirps[j].Id
	})

	chirpsJSON, err := json.Marshal(chirps)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	w.Write(chirpsJSON)

}
