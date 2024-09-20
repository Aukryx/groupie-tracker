package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/tlambert/groupie-tracker/static/handlers"
)

const port = ":8080"

func main() {
	http.HandleFunc("/", handlers.Home)
	http.HandleFunc("/artist/", handlers.ArtistPage)
	http.HandleFunc("/search", handlers.SearchArtists)
	http.HandleFunc("/filtered-results", handlers.FilteredResults)
	http.HandleFunc("/locations-autocomplete", handlers.LocationsAutocomplete)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	fmt.Println("(http://localhost:8080) - Server started on port", port)
	server := &http.Server{
		Addr:              port,              //adresse du server (le port choisi est à titre d'exemple)
		ReadHeaderTimeout: 10 * time.Second,  // temps autorisé pour lire les headers
		WriteTimeout:      10 * time.Second,  // temps maximum d'écriture de la réponse
		IdleTimeout:       120 * time.Second, // temps maximum entre deux rêquetes
		MaxHeaderBytes:    1 << 20,           // 1 MB // maxinmum de bytes que le serveur va lire
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
