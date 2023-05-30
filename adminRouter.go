package main

import (
	"net/http"
	"html/template"
	
)

func  metricsHandler (w http.ResponseWriter, r *http.Request, apiCfg *apiConfig){
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
}