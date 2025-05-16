# URL Shortener Service

A simple, fast URL shortener service written in Go that generates shortened URLs and provides redirect functionality.

## Features

- Generate shortened URLs from long URLs
- Redirect from shortened URLs to original URLs
- Persistent storage using SQLite database
- RESTful API for URL shortening
- SHA-256 based URL shortening algorithm

## Requirements

- Go 1.14 or higher
- SQLite3

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/SUSHANT-0210/URL_Shortener.git
   ```

2. Install dependencies:
   ```bash
   go get github.com/mattn/go-sqlite3
   ```

3. Build the application:
   ```bash
   go build -o url-shortener
   ```

## Usage

### Starting the Server

Run the compiled binary:

```bash
./url-shortener
```

The server will start on port 8080.

### API Endpoints

#### Shorten URL

```
POST /shorten
```

Request body:
```json
{
  "url": "https://example.com/very/long/url/that/needs/shortening"
}
```

Response:
```json
{
  "short_url": "a1b2c3d4"
}
```

#### Redirect to Original URL

```
POST /redirect/{shortURL}
```

Example:
```
POST /redirect/a1b2c3d4
```

This will redirect to the original URL associated with the shortened URL.

## How It Works

1. The service generates a short URL by:
   - Taking the SHA-256 hash of the original URL
   - Using the first 8 characters of the hash as the short URL ID

2. URLs are stored in an SQLite database with the following schema:
   ```sql
   CREATE TABLE urls (
       id TEXT PRIMARY KEY,
       original_url TEXT,
       short_url TEXT,
       created_at DATETIME
   )
   ```

3. When a user accesses a shortened URL, the service looks up the original URL in the database and redirects the user.

## Error Handling

- If an invalid URL is provided, an appropriate error message is returned.
- If a shortened URL is not found in the database, a 404 error is returned.
- If there are database errors, a 500 error is returned.

## Security Considerations

- This service does not implement any authentication or rate limiting.
- For production use, consider adding:
  - Rate limiting to prevent abuse
  - Authentication for URL creation
  - HTTPS support for secure data transmission
