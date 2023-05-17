package main

import (
	"fmt"
	"log"
	"net/http"
)

type apiConfig struct {
	fileserverHits int
}

func main() {
	const port = "8080"
	mux := http.NewServeMux()      // it is a HTTP request multiplexer.
	corsMux := middlewareCors(mux) // it wraps mux with the middlewareCors, this ensure that all request first passes through CORS MIDDLEWARE

	var srv http.Server   // we create a variable srv which is of type http.Server !
	srv.Addr = ":" + port // corrected server address
	srv.Handler = corsMux // the handler will be corsMux , it is to ensure that every request must first go through this middleware
	var apiCfg apiConfig

	fileHandler := http.FileServer(http.Dir("."))

	mux.Handle("/", apiCfg.middlewareMetricsInc(fileHandler)) // before going to homepage, first it will go to middlewareMetricsInx
	mux.Handle("/metrics", apiCfg.metricsHandler())
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8") // setting the content type
		w.WriteHeader(http.StatusOK)                                // setting the status code to 200
		w.Write([]byte(http.StatusText(http.StatusOK)))             // setting the body message
	})

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())
}

func middlewareCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}


func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits += 1
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) metricsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Hits: %v", cfg.fileserverHits)
	})
}
