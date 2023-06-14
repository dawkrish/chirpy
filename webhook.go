package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

func webhook(w http.ResponseWriter, r *http.Request, db *DB, apiCfg *apiConfig){
	authorizationString := r.Header.Get("Authorization")
	apiKey := strings.TrimPrefix(authorizationString, "ApiKey ")
	if apiKey != apiCfg.polkaKey {
		w.WriteHeader(401)
		w.Write([]byte("{}"))
		return
	}

	type requestBody struct{
		Event string `json:"event"`
		Data map[string] int `json:"data"`
	}
	bodyFetched := requestBody{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&bodyFetched)

	if err != nil {
		http.Error(w,err.Error(),400)
		return 
	}
	if bodyFetched.Event != "user.upgraded" {
		w.WriteHeader(200)
		w.Write([]byte("{}"))
		return
	}
	if bodyFetched.Event == "user.upgraded"{
		err := ChirpyRed(db, bodyFetched.Data["user_id"])
		if err != nil {
			http.Error(w,err.Error(), 404)
			w.Write([]byte("{}"))
			return
		}else{
			w.WriteHeader(200)
			w.Write([]byte("{}"))
		}
	}

}