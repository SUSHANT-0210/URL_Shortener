package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/time/rate"
)

// server-side private key
var jwtSecret = []byte(getEnvOrDefault("JWT_Secret", "!@#$%^&*()_+"))

// Structure to hold in db
type URL struct {
	Id          string    `json:"id"`
	OriginalURL string    `json:"original_url"`
	ShortURL    string    `json:"short_url"`
	CreatedAt   time.Time `json:"created_at"`
	UserId      string    `json:"user_id"`
}

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Password string `json:"-"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token   string `json:"token"`
	Message string `json:"message"`
}

type Claims struct {
	UserId   string `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
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

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func initializeDB() error {
	var err error
	db, err = sql.Open("sqlite3", "./url_shortner.db")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	urlQuery := `
    CREATE TABLE IF NOT EXISTS urls (
        id TEXT PRIMARY KEY,
        original_url TEXT,
        short_url TEXT,
        created_at DATETIME,
        user_id TEXT,
        FOREIGN KEY (user_id) REFERENCES users (id)
    )`
	_, err = db.Exec(urlQuery)
	if err != nil {
		return fmt.Errorf("failed to create urls table: %w", err)
	}

	userQuery := `
    CREATE TABLE IF NOT EXISTS users (
        id TEXT PRIMARY KEY,
        username TEXT UNIQUE,
        password TEXT
    )`
	_, err_ := db.Exec(userQuery)
	if err_ != nil {
		return fmt.Errorf("failed to create users table: %w", err)
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

func generateUserID(username string) string {
	hasher := sha256.New()
	hasher.Write([]byte(username))
	data := hasher.Sum((nil))
	hash := fmt.Sprintf("%x", data)
	return hash[:16]
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func generateJWT(userID, username string) (string, error) {

	expirationDate := time.Now().Add(24 * time.Hour)
	claims := Claims{
		UserId:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationDate),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func validateJWT(tokenString string) (*Claims, error) {

	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			http.Error(w, "Bearer token required", http.StatusUnauthorized)
			return
		}
		claims, err := validateJWT(tokenString)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		r.Header.Set("X-User-ID", claims.UserId)
		r.Header.Set("X-Username", claims.Username)

		next(w, r)
	}
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

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Methos not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		http.Error(w, "Username and password required", http.StatusBadRequest)
		return
	}

	var existingUser string

	checkQuery := `SELECT id FROM users WHERE username = ?`
	err_ := db.QueryRow(checkQuery, req.Username).Scan(&existingUser)

	if err_ != nil && err_ != sql.ErrNoRows {
		// Handle other SQL errors
		fmt.Printf("Error executing query: %v\n", err_)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	hashedPassword, err__ := hashPassword(req.Password)
	if err__ != nil {
		http.Error(w, "Error processing password", http.StatusInternalServerError)
		return
	}

	userID := generateUserID(req.Username)

	insertQuery := `INSERT INTO users (id, username, password) VALUES (?, ?, ?)`
	_, err = db.Exec(insertQuery, userID, req.Username, hashedPassword)
	if err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	token, err := generateJWT(userID, req.Username)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	response := AuthResponse{
		Token:   token,
		Message: "User registered successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	var user User

	query := `SELECT id, username, password FROM users WHERE username = ?`
	err = db.QueryRow(query, req.Username).Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	if !checkPasswordHash(req.Password, user.Password) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := generateJWT(user.ID, user.Username)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	response := AuthResponse{
		Token:   token,
		Message: "Login successful",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func storeInfoInDB(originalURL string, userID string) error {
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
	fmt.Printf("Storing URL: %s for user: %s\n", originalURL, userID)
	fmt.Printf("Short URL: %s\n", shortURL)

	// Since the shortURL is a unique string generated for each original URL, therefore it is serving as the id too.
	queryInsert := `INSERT INTO urls (id, original_url, short_url, created_at, user_id) VALUES (?, ?, ?, ?, ?)`
	_, err = db.Exec(queryInsert, id, originalURL, shortURL, time.Now(), userID)
	if err != nil {
		fmt.Printf("Error inserting URL: %v\n", err)
		return err
	}

	fmt.Println("URL stored successfully.")
	return nil
}

func getURLStructure(id string) (URL, error) {
	query := `
	SELECT id, original_url, short_url, created_at, user_id
	FROM urls
	WHERE id = ?
	`
	row := db.QueryRow(query, id)
	fmt.Printf("Row %v", row)
	var url URL
	err := row.Scan(&url.Id, &url.OriginalURL, &url.ShortURL, &url.CreatedAt, &url.UserId)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return URL{}, errors.New("URL not found")
		}
		return URL{}, err
	}
	return url, nil
}

func getUserURLs(userID string) ([]URL, error) {
	query := `
	SELECT id, original_url, short_url, created_at, user_id
	FROM urls
	WHERE user_id = ?
	ORDER BY created_at DESC
	`
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var urls []URL

	for rows.Next() {
		var url URL
		err := rows.Scan(&url.Id, &url.OriginalURL, &url.ShortURL, &url.CreatedAt, &url.UserId)
		if err != nil {
			return nil, err
		}
		urls = append(urls, url)

	}
	return urls, nil
}

func RootPageHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "URL Shortener API\n\nEndpoints:\n")
	fmt.Fprintf(w, "POST /register - Register a new user\n")
	fmt.Fprintf(w, "POST /login - Login user\n")
	fmt.Fprintf(w, "POST /shorten - Shorten URL (requires auth)\n")
	fmt.Fprintf(w, "GET /redirect/{id} - Redirect to original URL\n")
	fmt.Fprintf(w, "GET /urls - Get user's URLs (requires auth)\n")
}

func shortURLHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := r.Header.Get("X-User-ID")
	// fmt.Printf("Header :- %s", r.Header)
	if userID == "" {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	var data struct {
		URL string `json:"url"`
	}

	fmt.Printf("Request received: %s\n", r.URL.Path)

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	if data.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	err__ := storeInfoInDB(data.URL, userID)

	if err__ != nil {
		http.Error(w, "Error storing URL in database", http.StatusInternalServerError)
		return
	}

	shortURL_ := generateShortURL(data.URL)
	response := struct {
		ShortURL string `json:"short_url"`
		ID       string `json:"id"`
	}{
		ShortURL: shortURL_,
		ID:       shortURL_,
	}

	w.Header().Set("Content-Type", "application/json")
	err_ := json.NewEncoder(w).Encode(response)
	if err_ != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func redirectURLHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/redirect/"):]
	if id == "" {
		http.Error(w, "ID required", http.StatusBadRequest)
		return
	}
	urlStructure, err := getURLStructure(id)
	if err != nil {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}
	http.Redirect(w, r, urlStructure.OriginalURL, http.StatusFound)
}

func getUserURLsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}
	urls, err := getUserURLs(userID)
	if err != nil {
		http.Error(w, "Error fetching URLs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(urls)

}

func main() {
	// Initialize the database
	fmt.Println("Starting URL Shortener with JWT Authentication...")

	err := initializeDB()
	if err != nil {
		fmt.Println("Error initializing database:", err)
		return
	}

	defer db.Close()

	// Public endpoints
	http.HandleFunc("/", RootPageHandler)
	http.HandleFunc("/register", rateLimiterMiddleware(registerHandler))
	http.HandleFunc("/login", rateLimiterMiddleware(loginHandler))
	http.HandleFunc("/redirect/", rateLimiterMiddleware(redirectURLHandler))

	// Protected endpoints (require authentication)
	http.HandleFunc("/shorten", rateLimiterMiddleware(authMiddleware(shortURLHandler)))
	http.HandleFunc("/urls", rateLimiterMiddleware(authMiddleware(getUserURLsHandler)))

	// Start HTTP server on port 8080
	fmt.Println("Server starting on port 8080...")
	fmt.Println("Available endpoints:")
	fmt.Println("  POST /register - Register new user")
	fmt.Println("  POST /login - Login user")
	fmt.Println("  POST /shorten - Shorten URL (requires Bearer token)")
	fmt.Println("  GET /redirect/{id} - Redirect to original URL")
	fmt.Println("  GET /urls - Get user's URLs (requires Bearer token)")

	err_ := http.ListenAndServe(":8080", nil)
	if err_ != nil {
		fmt.Println("Error starting server:", err)
		return
	}
}
