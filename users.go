package main

import (
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func userPost(w http.ResponseWriter, r *http.Request, db *DB) {
	type requestBodyParams struct {
		Password string `json:"password"`
		Email string `json:"email"`
	}
	decoder := json.NewDecoder(r.Body)
	bodyFetched := requestBodyParams{}
	err := decoder.Decode(&bodyFetched)

	// now the content of the request body is in bodyFetched variable !

	if err != nil {
		http.Error(w, "Something went wrong!", http.StatusBadRequest)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(bodyFetched.Password),bcrypt.DefaultCost)
	if err != nil {
		http.Error(w,err.Error(),http.StatusBadRequest)
	}

	user, err := db.CreateUser(bodyFetched.Email,string(hashedPassword))
	if err != nil {
		http.Error(w,err.Error(),http.StatusConflict)
		return
	}
	userWithoutPassword := struct{
		Id int
		Email string
	}{
		Id: user.Id,
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
