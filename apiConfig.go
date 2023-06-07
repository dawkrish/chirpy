package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

type apiConfig struct {
	fileserverHits int
	jwtSecret []byte
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
		log.Print("Something went wrong!")
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
		log.Print(err)
		log.Print("Something went wrong in response body!")
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
	log.Print("Header set")
	w.WriteHeader(201)
	log.Print("response set !")
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

// chirpsGetByID retrieves a chirp by its ID from the database and sends it as a response.
func chirpsGetById(w http.ResponseWriter, r *http.Request, db *DB) {
	// Extract the chirp ID from the URL parameter
	id := chi.URLParam(r, "chirpID")

	// Convert the ID to an integer
	numericID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid chirp ID", http.StatusBadRequest)
		return
	}

	// Retrieve all chirps from the database
	chirps, err := db.GetChirps()
	if err != nil {
		http.Error(w, "Database problem", http.StatusInternalServerError)
		return
	}

	// Find the chirp with the matching ID
	var foundChirp *Chirp
	for _, chirp := range chirps {
		if chirp.Id == numericID {
			foundChirp = &chirp
			break
		}
	}

	// If a chirp with the ID is found, marshal it to JSON and send it as the response
	if foundChirp != nil {
		responseJSON, err := json.MarshalIndent(foundChirp, "", "  ")
		if err != nil {
			http.Error(w, "Error marshalling JSON", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(responseJSON)
		return
	}

	// If no chirp with the ID is found, send a 404 Not Found response
	http.NotFound(w, r)
}
