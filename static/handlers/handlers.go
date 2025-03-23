package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/tlambert/groupie-tracker/features"
)

// Will render any template .html from the templates directory
func RenderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {

	funcMap := template.FuncMap{
		"json": func(v interface{}) template.JS {
			a, err := json.Marshal(v)
			if err != nil {
				log.Printf("Error marshaling JSON: %v", err)
				return template.JS("{}")
			}
			return template.JS(a)
		},
	}
	// include in map js func
	page, err := template.New(tmpl + ".html").Funcs(funcMap).ParseFiles("templates/" + tmpl + ".html")

	if err != nil {
		w.WriteHeader(404)
		http.Error(w, "error 404", http.StatusNotFound)
		log.Printf("error template %v", err)
		return
	}
	err = page.Execute(w, data)
	if err != nil {
		http.Error(w, "Error 500, Internal server error", http.StatusInternalServerError)
		log.Printf("error template %v", err)
		return
	}
}

// Page where all artists are displayed
func Home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		RenderTemplate(w, "error404", nil)
	} else {
		var artists []features.Artist
		urlAPI := "https://groupietrackers.herokuapp.com/api/artists"

		err := features.FetchData(urlAPI, &artists)
		if err != nil {
			features.Error(w, http.StatusInternalServerError, "Error fetching artist data")
			return
		}

		RenderTemplate(w, "index", artists)
	}
}

func ArtistPage(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(r.URL.Path, "/")
	if len(path) < 3 {
		features.Error(w, http.StatusBadRequest, "Invalid artist ID")
		return
	}
	artistID, err := strconv.Atoi(path[2])
	if err != nil {
		features.Error(w, http.StatusBadRequest, "Invalid artist ID")
		return
	}

	if artistID <= 0 || artistID > 52 {
		features.Error(w, http.StatusBadRequest, "Invalid artist ID")
		return
	}

	var pageData features.ArtistPageData

	// Fetch artist data
	artistURL := fmt.Sprintf("https://groupietrackers.herokuapp.com/api/artists/%d", artistID)
	err = features.FetchData(artistURL, &pageData.Artist)
	if err != nil {
		features.Error(w, http.StatusInternalServerError, "Error fetching artist data")
		return
	}

	// Fetch relations
	relationsURL := fmt.Sprintf("https://groupietrackers.herokuapp.com/api/relation/%d", artistID)
	err = features.FetchData(relationsURL, &pageData.Relations)
	if err != nil {
		features.Error(w, http.StatusInternalServerError, "Error fetching relations data")
		return
	}

	locations := make([]string, 0)
	for location := range pageData.Relations.DatesLocations {
		locations = append(locations, location)
	}
	pageData.ArtistLocations = locations

	RenderTemplate(w, "artist", pageData)
}

func SearchArtists(w http.ResponseWriter, r *http.Request) {
	query := strings.ToLower(r.URL.Query().Get("q"))

	var artists []features.Artist
	var locationsData struct {
		Index []struct {
			ID        int      `json:"id"`
			Locations []string `json:"locations"`
		} `json:"index"`
	}

	// Fetch artists data
	err := features.FetchData("https://groupietrackers.herokuapp.com/api/artists", &artists)
	if err != nil {
		features.Error(w, http.StatusInternalServerError, "Error fetching artist data")
		return
	}

	// Fetch locations data
	err = features.FetchData("https://groupietrackers.herokuapp.com/api/locations", &locationsData)
	if err != nil {
		features.Error(w, http.StatusInternalServerError, "Error fetching location data")
		return
	}

	suggestions := []features.SearchSuggestion{}

	for i, artist := range artists {
		// Check artist/band name
		if strings.HasPrefix(strings.ToLower(artist.Name), query) {
			suggestions = append(suggestions, features.SearchSuggestion{Value: artist.Name, Type: "artist/band", ID: artist.ID})
		}

		// Check creation date
		creationDateStr := strconv.Itoa(artist.CreationDate)
		if strings.HasPrefix(creationDateStr, query) {
			suggestions = append(suggestions, features.SearchSuggestion{Value: creationDateStr, Type: "creation date", ID: artist.ID})
		}

		// Check members (not visible on index page, but searchable)
		for _, member := range artist.Members {
			if strings.HasPrefix(strings.ToLower(member), query) {
				suggestions = append(suggestions, features.SearchSuggestion{Value: member, Type: "member", ID: artist.ID})
			}
		}

		// Check first album (not visible on index page, but searchable)
		if strings.HasPrefix(strings.ToLower(artist.FirstAlbum), query) {
			suggestions = append(suggestions, features.SearchSuggestion{Value: artist.FirstAlbum, Type: "first album", ID: artist.ID})
		}

		// Check locations (not visible on index page, but searchable)
		if i < len(locationsData.Index) {
			for _, location := range locationsData.Index[i].Locations {
				locationPays := strings.Split(location, "-")
				Pays := locationPays[1]
				if strings.HasPrefix(strings.ToLower(location), query) || strings.HasPrefix(strings.ToLower(Pays), query) {
					suggestions = append(suggestions, features.SearchSuggestion{Value: location, Type: "location", ID: artist.ID})
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(suggestions)
}

func FilteredResults(w http.ResponseWriter, r *http.Request) {
	// Parse filter parameters
	creationDateMin, _ := strconv.Atoi(r.URL.Query().Get("creationDateMin"))
	creationDateMax, _ := strconv.Atoi(r.URL.Query().Get("creationDateMax"))
	firstAlbumMin, _ := strconv.Atoi(r.URL.Query().Get("firstAlbumMin"))
	firstAlbumMax, _ := strconv.Atoi(r.URL.Query().Get("firstAlbumMax"))
	memberCounts := strings.Split(r.URL.Query().Get("memberCounts"), ",")
	locations := r.URL.Query().Get("locations")

	// Fetch all artists
	var allArtists []features.Artist
	err := features.FetchData("https://groupietrackers.herokuapp.com/api/artists", &allArtists)
	if err != nil {
		features.Error(w, http.StatusInternalServerError, "Error fetching artist data")
		return
	}

	// Filter artists
	filteredArtists := features.FilterArtists(allArtists, creationDateMin, creationDateMax, firstAlbumMin, firstAlbumMax, memberCounts, locations)

	// Render the template with filtered artists
	RenderTemplate(w, "index", filteredArtists)
}

func LocationsAutocomplete(w http.ResponseWriter, r *http.Request) {
	locations, err := features.GetAllUniqueLocations()
	if err != nil {
		http.Error(w, "Error fetching locations", http.StatusInternalServerError)
		return
	}

	query := strings.ToLower(r.URL.Query().Get("q"))
	var matches []string
	for _, location := range locations {
		if strings.HasPrefix(strings.ToLower(location), query) {
			matches = append(matches, location)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(matches)
}
