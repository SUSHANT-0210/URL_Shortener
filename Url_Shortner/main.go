package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"golang.org/x/time/rate"
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

var db *sql.DB
var limiter = rate.NewLimiter(1, 5)

func initializeDB() error {
	var err error
	db, err = sql.Open("sqlite3", "./url_shortner.db")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	query := `
    CREATE TABLE IF NOT EXISTS urls (
        id TEXT PRIMARY KEY,
        original_url TEXT,
        short_url TEXT,
        created_at DATETIME
    )`
	_, err = db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}
	return nil
}

func generateShortURL(original_url string) string {
	hasher := sha256.New()
	hasher.Write([]byte(original_url))
	data := hasher.Sum(nil)
	hash := fmt.Sprintf("%x", data)
	return hash[:8]
}

func storeInfoInDB(originalURL string) error {
	shortURL := generateShortURL(originalURL)
	id := shortURL

	// Check if the URL already exists
	queryCheck := `SELECT id FROM urls WHERE original_url = ?`
	var existingID string
	err := db.QueryRow(queryCheck, originalURL).Scan(&existingID)
	if err == nil {
		// URL already exists
		fmt.Printf("URL already exists with ID: %s\n", existingID)
		return nil
	} else if err != sql.ErrNoRows {
		// An actual error occurred
		fmt.Printf("Error checking for existing URL: %v\n", err)
		return err
	}

	// URL does not exist, insert it into the database
	fmt.Printf("Storing URL: %s\n", originalURL)
	fmt.Printf("Short URL: %s\n", shortURL)

	queryInsert := `INSERT INTO urls (id, original_url, short_url, created_at) VALUES (?, ?, ?, ?)`
	_, err = db.Exec(queryInsert, id, originalURL, shortURL, time.Now())
	if err != nil {
		fmt.Printf("Error inserting URL: %v\n", err)
		return err
	}

	fmt.Println("URL stored successfully.")
	return nil
}

func getURLStructure(id string) (URL, error) {
	query := `
	SELECT id, original_url, short_url, created_at
	FROM urls
	WHERE id = ?
	`
	row := db.QueryRow(query, id)

	var url URL
	err := row.Scan(&url.Id, &url.OriginalURL, &url.ShortURL, &url.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return URL{}, errors.New("URL not found")
		}
		return URL{}, err
	}
	return url, nil
}

func rateLimiterMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			fmt.Println("Rate limit exceeded for client")
			return
		}
		next(w, r)
	}
}

func RootPageHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Invalid URL!\n")
}

func shortURLHandler(w http.ResponseWriter, r *http.Request) {
	var data struct {
		URL string `json:"url"`
	}

	fmt.Printf("Request received: %s\n", r.URL.Path)

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	err__ := storeInfoInDB(data.URL)

	if err__ != nil {
		http.Error(w, "Error storing URL in database", http.StatusInternalServerError)
	}

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
	// Initialize the database
	fmt.Println("Starting program...")

	err := initializeDB()
	if err != nil {
		fmt.Println("Error initializing database:", err)
		return
	}

	defer db.Close()

	http.HandleFunc("/shorten", rateLimiterMiddleware(shortURLHandler))
	http.HandleFunc("/redirect/", rateLimiterMiddleware(redirectURLHandler))
	http.HandleFunc("/", RootPageHandler)
	// Start HTTP server on port 8080
	fmt.Println("Starting server on port 8080...")
	err_ := http.ListenAndServe(":8080", nil)
	if err_ != nil {
		fmt.Println("Error starting server:", err)
		return
	}
}
