package main

import (
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func userLogin(w http.ResponseWriter, r *http.Request, db *DB) {
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

	// first fetch the user with email, then compare the password !
	findUser, err := GetUser(db, bodyFetched.Email)

	if err != nil {
		http.Error(w, "user not found", http.StatusBadRequest)
	}

	err = bcrypt.CompareHashAndPassword([]byte(findUser.Password), []byte(bodyFetched.Password))

	if err == nil {
		// correct password entered ! has been found !
		// write response !
		userWithoutPassword := struct {
			Id    int
			Email string
		}{
			Id:    findUser.Id,
			Email: findUser.Email,
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
		w.WriteHeader(200)
		w.Write(responseJSON)
	}else{
		http.Error(w, "password does not match !", http.StatusUnauthorized)
	}
}
