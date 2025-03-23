package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/tlambert/groupie-tracker/static/handlers"
)

func main() {
	// Get port from environment variable (for Render) or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Add colon prefix if not present
	if port[0] != ':' {
		port = ":" + port
	}

	http.HandleFunc("/", handlers.Home)
	http.HandleFunc("/artist/", handlers.ArtistPage)
	http.HandleFunc("/search", handlers.SearchArtists)
	http.HandleFunc("/filtered-results", handlers.FilteredResults)
	http.HandleFunc("/locations-autocomplete", handlers.LocationsAutocomplete)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("Server started on port", port)

	server := &http.Server{
		Addr:              port,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
