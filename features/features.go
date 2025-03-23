package features

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"text/template"
)

// Every structures that datas from JSON files will be put into

type Artist struct {
	ID           int      `json:"id"`
	Image        string   `json:"image"`
	Name         string   `json:"name"`
	Members      []string `json:"members"`
	CreationDate int      `json:"creationDate"`
	FirstAlbum   string   `json:"firstAlbum"`
}

type ArtistLocation struct {
	Index []struct {
		ID        int      `json:"id"`
		Locations []string `json:"locations"`
	} `json:"index"`
}

type ArtistDates struct {
	Dates []string `json:"dates"`
}

type ArtistRelation struct {
	DatesLocations map[string][]string `json:"datesLocations"`
}

type ArtistPageData struct {
	Artist          Artist
	Locations       ArtistLocation
	Dates           ArtistDates
	Relations       ArtistRelation
	ArtistLocations []string
}

type SearchSuggestion struct {
	Value string `json:"value"`
	Type  string `json:"type"`
	ID    int    `json:"id"`
}

// Iterates through all artists and append into a new tab of strings, according to the filter values used
func FilterArtists(artists []Artist, creationDateMin, creationDateMax, firstAlbumMin, firstAlbumMax int, memberCounts []string, locations string) []Artist {
	var filtered []Artist
	var locationsData ArtistLocation
	err := FetchData("https://groupietrackers.herokuapp.com/api/locations", &locationsData)
	if err != nil {
		log.Printf("Error fetching locations: %v", err)
		return filtered
	}

	for _, artist := range artists {
		if artist.CreationDate < creationDateMin || artist.CreationDate > creationDateMax {
			continue
		}

		firstAlbumYear, _ := strconv.Atoi(strings.Split(artist.FirstAlbum, "-")[2])
		if firstAlbumYear < firstAlbumMin || firstAlbumYear > firstAlbumMax {
			continue
		}

		if !Contains(memberCounts, strconv.Itoa(len(artist.Members))) {
			continue
		}

		//handle the use of location filter
		if locations != "" {
			artistLocations := GetArtistLocations(locationsData, artist.ID)
			if !ContainsLocation(artistLocations, locations) {
				continue
			}
		}

		filtered = append(filtered, artist)
	}

	return filtered
}

// Will Fetch Data from the JSON file
func FetchData(url string, target interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("HTTP request error: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Unexpected status code: %d", resp.StatusCode)
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(target)
	if err != nil {
		log.Printf("JSON decode error: %v", err)
		return err
	}

	return nil
}

func Error(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	var tmpl string
	switch status {
	case http.StatusBadRequest:
		tmpl = "error400"
	case http.StatusNotFound:
		tmpl = "error404"
	case http.StatusInternalServerError:
		tmpl = "error500"
	default:
		tmpl = "unexpected_error"
		log.Printf("Unexpected error status: %d", status)
	}
	page, err := template.ParseFiles("templates/" + tmpl + ".html")
	if err != nil {
		http.Error(w, "Server Error", http.StatusInternalServerError)
		log.Printf("Error parsing template: %v", err)
		return
	}
	data := struct {
		Status  int
		Message string
	}{
		Status:  status,
		Message: message,
	}
	err = page.Execute(w, data)
	if err != nil {
		http.Error(w, "Server Error", http.StatusInternalServerError)
		log.Printf("Error executing template: %v", err)
	}
}

// LOCATIONS \\

// Iterates through all of the locations API and verify if the Artist ID is the same in both APIs
func GetArtistLocations(locationsData ArtistLocation, artistID int) []string {
	for _, item := range locationsData.Index {
		if item.ID == artistID {
			return item.Locations
		}
	}
	return nil
}

func ContainsLocation(locations []string, query string) bool {
	query = strings.ToLower(query)
	for _, location := range locations {
		if strings.Contains(strings.ToLower(location), query) {
			return true
		}
	}
	return false
}

func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func GetAllUniqueLocations() ([]string, error) {
	var locationsData ArtistLocation
	err := FetchData("https://groupietrackers.herokuapp.com/api/locations", &locationsData)
	if err != nil {
		return nil, err
	}

	uniqueLocations := make(map[string]bool)
	for _, item := range locationsData.Index {
		for _, location := range item.Locations {
			uniqueLocations[location] = true
		}
	}

	var result []string
	for location := range uniqueLocations {
		result = append(result, location)
	}

	return result, nil
}
