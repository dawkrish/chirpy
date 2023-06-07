package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
)

func refresh(w http.ResponseWriter, r *http.Request, db *DB, apiCfg *apiConfig) {
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

	if !ok || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}
	if claims.Issuer == "chirpy-access" {
		http.Error(w, "no access", http.StatusUnauthorized)
		return
	}

	if GetRevoke(db, tokenString) {
		http.Error(w, "this token is revoked !", http.StatusUnauthorized)
		return
	}

	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Subject:   claims.Subject,
		ExpiresAt: time.Now().UTC().Unix() + int64(time.Hour),
		Issuer:   "chirpy-access",
		IssuedAt: time.Now().UTC().Unix(),
	})

	signedNewToken, _ := newToken.SignedString(apiCfg.jwtSecret)

	// Create the response shape
	response := struct {
		Token string `json:"token"`
	}{
		Token: signedNewToken,
	}

	// Convert the response to JSON
	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set the response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// Write the response body
	w.Write(responseJSON)
}

func revoke(w http.ResponseWriter, r *http.Request, db *DB, apiCfg *apiConfig) {
	authorizationString := r.Header.Get("Authorization")
	tokenString := strings.TrimPrefix(authorizationString, "Bearer ")

	err := AddRevoke(db, tokenString)
	if err != nil {
		log.Print("Error occurred in revoke handler: " + err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	emptyJSON := struct{}{}
	json.NewEncoder(w).Encode(emptyJSON)
}
