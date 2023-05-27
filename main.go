package main

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"html/template"
	"log"
	"net/http"
	"strings"
	"sort"
)

type Chirp struct {
	Id   int    `json:"id"`
	Body string `json:"body"`
}

func main() {
	const port = "8080"
	DB, err := NewDB("")
	if err != nil {
		return
	}
	r := chi.NewRouter()
	apiRouter := chi.NewRouter()
	adminRouter := chi.NewRouter()
	corsMux := middlewareCors(r) // it wraps mux with the middlewareCors, this ensure that all request first passes through CORS MIDDLEWARE

	var srv http.Server   // we create a variable srv which is of type http.Server !
	srv.Addr = ":" + port // corrected server address
	srv.Handler = corsMux // the handler will be corsMux , it is to ensure that every request must first go through this middleware

	var apiCfg apiConfig
	apiCfg.fileserverHits = 0

	fileHandler := http.FileServer(http.Dir("."))

	r.Mount("/", apiCfg.middlewareMetricsInc(fileHandler))
	apiRouter.Get("/metrics", apiCfg.metricsHandler)
	apiRouter.Get("/healthz", handlerReadiness)

	apiRouter.Post("/chirps", func(w http.ResponseWriter, r *http.Request) {
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
		responseBody, err := DB.CreateChirp(cleanedChirpStr)
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

	})

	apiRouter.Get("/chirps", func(w http.ResponseWriter, r *http.Request) {
		chirps, err := DB.GetChirps()
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
		
	})

	adminRouter.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
		data := struct { // The data that we are sending to html file !
			Hits int
		}{
			Hits: apiCfg.fileserverHits,
		}

		tmpl, err := template.ParseFiles("assets/metrics.html") // we parse the html file in tmpl !!
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		err = tmpl.Execute(w, data) // we write the data in tmpl !
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	})

	r.Mount("/api", apiRouter)
	r.Mount("/admin", adminRouter)

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())
}
