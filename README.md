# URL Shortener Service

A high-performance, Go-based URL shortener service that generates and manages shortened URLs while ensuring seamless redirection functionality. Designed for simplicity, scalability, and ease of use.

## Features

* **Efficient URL Shortening:** Shortens long URLs using a robust SHA-256-based algorithm.
* **Redirection Support:** Redirects shortened URLs to their original counterparts.
* **Persistent Storage:** Uses SQLite for reliable and persistent URL storage.
* **RESTful API:** Provides a clean API for URL shortening and redirection.
* **Rate Limiting:** Prevents abuse of the service through configurable rate-limiting mechanisms, here Token bucket implementation is used.
* **Error Handling:** Comprehensive error responses for invalid requests and system issues.

## Requirements

* **Go:** Version 1.14 or higher
* **SQLite:** SQLite3 for database management

## Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/SUSHANT-0210/URL_Shortener.git
   ```

2. Install dependencies:

   ```bash
   go mod tidy
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

The server will start on `http://localhost:8080`.

### API Endpoints

#### Shorten URL

**Endpoint:**
`POST /shorten`

**Request Body:**

```json
{
  "url": "https://example.com/very/long/url/that/needs/shortening"
}
```

**Response:**

```json
{
  "short_url": "a1b2c3d4"
}
```

#### Redirect to Original URL

**Endpoint:**
`POST /redirect/{shortURL}`

Example:

```
POST /redirect/a1b2c3d4
```

The service redirects to the original URL associated with `a1b2c3d4`.

## How It Works

1. **Short URL Generation:**

   * The SHA-256 hash of the original URL is computed.
   * The first 8 characters of the hash are used as the short URL ID.

2. **Database Management:**

   * SQLite database stores original URLs and their corresponding short URLs with timestamps.
   * Schema:

     ```sql
     CREATE TABLE urls (
         id TEXT PRIMARY KEY,
         original_url TEXT NOT NULL,
         short_url TEXT NOT NULL,
         created_at DATETIME DEFAULT CURRENT_TIMESTAMP
     );
     ```

3. **Request Handling:**

   * Short URL requests are processed via a REST API.
   * Redirect requests are resolved from the database and handled efficiently.

## Advanced Features

* **Rate Limiting:** Configurable rate limiter to prevent request abuse.
* **Error Handling:**

  * **400 Bad Request:** Invalid input.
  * **404 Not Found:** Short URL does not exist.
  * **500 Internal Server Error:** System/database errors.
* **Concurrency Support:** Handles multiple simultaneous requests efficiently.


## Future Enhancements

* Implementing authentication for secure URL shortening
* Support for customizable expiration of shortened URLs
