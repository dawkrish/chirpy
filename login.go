package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

func userLogin(w http.ResponseWriter, r *http.Request, db *DB, apiCfg *apiConfig) {
	type requestBodyParams struct {
		Password           string `json:"password"`
		Email              string `json:"email"`
		Expires_In_Seconds int    `json:"expires_in_seconds"`
	}

	decoder := json.NewDecoder(r.Body)
	bodyFetched := requestBodyParams{}
	err := decoder.Decode(&bodyFetched)

	// now the content of the request body is in bodyFetched variable !
	if err != nil {
		http.Error(w, "Something went wrong!", http.StatusBadRequest)
	}

	// first fetch the user with email, then compare the password !
	findUser, err := GetUser(db, bodyFetched.Email)

	if err != nil {
		http.Error(w, "user not found", http.StatusBadRequest)
	}

	err = bcrypt.CompareHashAndPassword([]byte(findUser.Password), []byte(bodyFetched.Password))

	if err == nil {

		var accessTokenExpiration int64 = int64(time.Hour)
		var refreshTokenExpiration int64 = int64(time.Hour * 24 * 60)

		accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
			Issuer:    "chirpy-access",
			IssuedAt:  time.Now().UTC().Unix(),
			ExpiresAt: accessTokenExpiration,
			Subject:   strconv.Itoa(findUser.Id),
		})
		accessTokenString, err := accessToken.SignedString(apiCfg.jwtSecret)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
			Issuer:    "chirpy-refresh",
			IssuedAt:  time.Now().UTC().Unix(),
			ExpiresAt: refreshTokenExpiration,
			Subject:   strconv.Itoa(findUser.Id),
		})
		refreshTokenString, err := refreshToken.SignedString(apiCfg.jwtSecret)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		// correct password entered ! has been found !
		// write response !
		userWithoutPassword := struct {
			Id           int    `json:"id"`
			Email        string `json:"email"`
			Token  string `json:"token"`
			RefreshToken string `json:"refresh_token"`
		}{
			Id:           findUser.Id,
			Email:        findUser.Email,
			Token:  accessTokenString,
			RefreshToken: refreshTokenString,
		}
		if err != nil {
			return
		}

		// Marshal the response into JSON
		responseJSON, err := json.Marshal(userWithoutPassword)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Set the response headers and write the response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(responseJSON)

	} else {
		http.Error(w, "password does not match !", http.StatusUnauthorized)
	}
}
