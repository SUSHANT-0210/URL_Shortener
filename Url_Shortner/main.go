package main

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// Structure to hold in db
type URL struct {
	Id          string    `json:"id"`
	OriginalURL string    `json:"original_url"`
	ShortURL    string    `json:"short_url"`
	CreatedAt   time.Time `json:"created_at"`
}

// Mapping in db
/*
	short_url ->{
		id: string,
		original_url: string,
		short_url: string,
		created_at: time.now(),
		}
*/

var urlDb = make(map[string]URL)

func generateShortURL(original_url string) string {
	hasher := sha256.New()
	hasher.Write([]byte(original_url))
	data := hasher.Sum(nil)
	hash := fmt.Sprintf("%x", data)
	return hash[:8]
}

func storeInfoInDB(originalURL string) {
	shortURL := generateShortURL(originalURL)
	id := shortURL
	urlDb[id] = URL{
		Id:          id,
		OriginalURL: originalURL,
		ShortURL:    shortURL,
		CreatedAt:   time.Now(),
	}
}

func getURLStructure(id string) (URL, error) {
	url, exists := urlDb[id]
	if !exists {
		return URL{}, errors.New("URL not found")
	}
	return url, nil
}

func RootPageHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Invalid URL!\n")
}

func shortURLHandler(w http.ResponseWriter, r *http.Request) {
	var data struct {
		URL string `json:"url"`
	}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	storeInfoInDB(data.URL)
	shortURL_ := generateShortURL(data.URL)
	response := struct {
		ShortURL string `json:"short_url"`
	}{ShortURL: shortURL_}

	w.Header().Set("Content-Type", "application/json")
	err_ := json.NewEncoder(w).Encode(response)
	if err_ != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func redirectURLHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/redirect/"):]
	urlStructure, err := getURLStructure(id)
	if err != nil {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}
	http.Redirect(w, r, urlStructure.OriginalURL, http.StatusFound)
}

func main() {

	http.HandleFunc("/shorten", shortURLHandler)
	http.HandleFunc("/redirect/", redirectURLHandler)
	http.HandleFunc("/", RootPageHandler)
	// Start HTTP server on port 8080
	fmt.Println("Starting server on port 8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
}
