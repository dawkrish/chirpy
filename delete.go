package main

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt"
)

func delete(w http.ResponseWriter, r *http.Request, db *DB, apiCfg *apiConfig) {
	chirpID := chi.URLParam(r, "chirpID")
	numericId, err := strconv.Atoi(chirpID)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	authorizationString := r.Header.Get("Authorization")
	tokenString := strings.TrimPrefix(authorizationString, "Bearer ")

	token, err := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(t *jwt.Token) (interface{}, error) {
		return apiCfg.jwtSecret, nil
	})
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(*jwt.StandardClaims)
	if !ok || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	claimsAuthorId, err := strconv.Atoi(claims.Subject)
	if err != nil {
		return
	}

	chirps, err := db.GetChirps()
	if err != nil {
		return
	}
	if numericId != claimsAuthorId {
		http.Error(w, "not authorized forbidden", http.StatusForbidden)
		return
	}

	log.Print("beginning of for loop")
	foundChirp := false
	
	log.Print(chirps)
	for _,val := range chirps {
		log.Print(val, numericId)
		if val.Id == numericId {
			log.Print("chirp found !")
			foundChirp = true
			err := DeleteChirp(numericId, db)
			if err != nil {
				log.Print(err.Error())
				http.Error(w, err.Error(), http.StatusForbidden)
				return
			}
		}
	}

	if !foundChirp {
		http.Error(w, "", http.StatusForbidden)
		log.Print("Chirp to delete not found")
	}

	w.Header().Set("Content-Type", "application/json")

	// Set the response status code
	w.WriteHeader(http.StatusOK)

	// Write an empty JSON object as the response body
	w.Write([]byte("{}"))
}
