package main

import (
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

type Chirp struct {
	Id   int    `json:"id"`
	Body string `json:"body"`
}

type User struct {
	Id   int    `json:"id"`
	Email string `json:"email"`
	Password string `json:"password"`
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

	apiRouter.Get("/chirps", func(w http.ResponseWriter, r *http.Request) {
		chirpsGet(w, r, DB)
	})

	apiRouter.Post("/chirps", func(w http.ResponseWriter, r *http.Request) {
		chirpsPost(w, r, DB)
	})

	apiRouter.Get("/chirps/{chirpID}", func(w http.ResponseWriter, r *http.Request) {
		chirpsGetById(w, r, DB)
	})

	apiRouter.Post("/users",func(w http.ResponseWriter, r *http.Request) {
		userPost(w,r,DB)
	})

	apiRouter.Post("/login",func(w http.ResponseWriter, r *http.Request) {
		userLogin(w,r,DB)
	})

	adminRouter.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
		metricsHandler(w,r,&apiCfg)
	})


	r.Mount("/api", apiRouter)
	r.Mount("/admin", adminRouter)

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())
}
