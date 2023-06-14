package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

type apiConfig struct {
	fileserverHits int
	jwtSecret      []byte
	polkaKey       string
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

func chirpsPost(w http.ResponseWriter, r *http.Request, db *DB, apiCfg *apiConfig) {
	authorizationString := r.Header.Get("Authorization")
	tokenString := strings.TrimPrefix(authorizationString, "Bearer ")
	// Validate and parse the JWT token
	token, err := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(t *jwt.Token) (interface{}, error) {
		return apiCfg.jwtSecret, nil
	})
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(*jwt.StandardClaims)

	userId := claims.Subject

	numericId, err := strconv.Atoi(userId)
	if err != nil {
		return
	}

	if !ok || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	type requestBodyParams struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	bodyFetched := requestBodyParams{}
	err = decoder.Decode(&bodyFetched)

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
	responseBody, err := db.CreateChirp(cleanedChirpStr, numericId)
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
	author_id := r.URL.Query().Get("author_id")
	sorting := r.URL.Query().Get("sort")

	if len(author_id) == 0 && sorting != "desc" ||
		len(author_id) == 0 && len(sorting) == 0 ||
		len(author_id) == 0 && sorting == "asc" {
		// return all chirps
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(chirpsJSON)
		return
	}
	if len(author_id) == 0 && sorting == "desc" {
		// return all chirps
		chirps, err := db.GetChirps()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Sort chirps by id in ascending order
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].Id > chirps[j].Id
		})
		chirpsJSON, err := json.Marshal(chirps)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(chirpsJSON)
		return
	}

	if len(author_id) != 0 && sorting == "asc" {
		//return specific chirp
		numericId, err := strconv.Atoi(author_id)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		chirps, err := db.GetChirps()
		if err != nil {
			http.Error(w, err.Error(), 400)
		}
		var ourChirps []Chirp
		for _, val := range chirps {
			if val.AuthorId == numericId {
				ourChirps = append(ourChirps, val)
			}
		}

		responseJson, err := json.Marshal(ourChirps)
		if err != nil {
			http.Error(w, err.Error(), 400)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(responseJson)
		return
	}

	if len(author_id) != 0 && sorting == "desc" {
		numericId, err := strconv.Atoi(author_id)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		chirps, err := db.GetChirps()
		if err != nil {
			http.Error(w, err.Error(), 400)
		}
		var ourChirps []Chirp
		for _, val := range chirps {
			if val.AuthorId == numericId {
				ourChirps = append(ourChirps, val)
			}
		}
		// reverse the order of ourChirps element !
		for i, val := range ourChirps {
			if i == len(ourChirps) / 2{
				break
			}
			temp := val
			ourChirps[i] = ourChirps[len(ourChirps)-1-i]
			ourChirps[len(ourChirps)-1-i] = temp
		}

		responseJson, err := json.Marshal(ourChirps)
		if err != nil {
			http.Error(w, err.Error(), 400)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(responseJson)
		return
	}

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
