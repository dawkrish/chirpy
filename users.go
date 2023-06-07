package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

func userPost(w http.ResponseWriter, r *http.Request, db *DB) {
	type requestBodyParams struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	decoder := json.NewDecoder(r.Body)
	bodyFetched := requestBodyParams{}
	err := decoder.Decode(&bodyFetched)

	// now the content of the request body is in bodyFetched variable !

	if err != nil {
		http.Error(w, "Something went wrong!", http.StatusBadRequest)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(bodyFetched.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	user, err := db.CreateUser(bodyFetched.Email, string(hashedPassword))
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	userWithoutPassword := struct {
		Id    int
		Email string
	}{
		Id:    user.Id,
		Email: user.Email,
	}

	// Marshal the response into JSON
	responseJSON, err := json.Marshal(userWithoutPassword)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Set the response headers and write the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(responseJSON)

}

func usersPut(w http.ResponseWriter, r *http.Request, db *DB, apiCfg *apiConfig) {
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

	// Check token expiration
	if time.Now().UTC().Unix() > claims.ExpiresAt {
		http.Error(w, "Token has expired", http.StatusUnauthorized)
		return
	}

	// Extract the user ID from the token claims
	userId, err := strconv.Atoi(claims.Subject)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Retrieve the user from the database using the user ID
	findUser, err := GetUserById(db, userId)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	type requestBodyParams struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	decoder := json.NewDecoder(r.Body)
	bodyFetched := requestBodyParams{}
	err = decoder.Decode(&bodyFetched)

	if err != nil {
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(bodyFetched.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	// now the content of the request body is in bodyFetched variable !

	if err != nil {
		http.Error(w, "Something went wrong!", http.StatusBadRequest)
	}

	updatedUser, err := UpdateUser(findUser.Id, bodyFetched.Email, string(hashedPassword), db)
	if err != nil {
		log.Fatal(err.Error() + " -> update error")
		return
	}

	userWithoutPassword := struct {
		Id    int
		Email string
	}{
		Id:    updatedUser.Id,
		Email: updatedUser.Email,
	}

	// Marshal the response into JSON
	responseJSON, err := json.Marshal(userWithoutPassword)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Set the response headers and write the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(responseJSON)
}
