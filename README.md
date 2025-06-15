# URL Shortener Service with JWT Authentication

A high-performance, Go-based URL shortener service with user authentication, JWT token management, and secure URL shortening capabilities. Features comprehensive user management, rate limiting, and persistent storage for scalable URL shortening operations.

## Features

* **User Authentication:** Complete user registration and login system with JWT tokens
* **Secure URL Shortening:** JWT-authenticated URL shortening with user-specific management
* **Efficient Algorithm:** SHA-256-based URL shortening generating 8-character unique identifiers  
* **Seamless Redirection:** Fast redirection from shortened URLs to original destinations
* **User Management:** Users can view and manage their own shortened URLs
* **Persistent Storage:** SQLite database for reliable data persistence with foreign key relationships
* **RESTful API:** Clean, well-structured API endpoints for all operations
* **Rate Limiting:** Token bucket implementation preventing service abuse (1 request/second, burst of 5)
* **Password Security:** Bcrypt hashing for secure password storage
* **Comprehensive Error Handling:** Detailed error responses for various failure scenarios
* **CORS Ready:** Structured for easy frontend integration

## Requirements

* **Go:** Version 1.14 or higher
* **SQLite:** SQLite3 for database management
* **Dependencies:** 
  - `github.com/golang-jwt/jwt/v5` - JWT token handling
  - `github.com/mattn/go-sqlite3` - SQLite driver
  - `golang.org/x/crypto/bcrypt` - Password hashing
  - `golang.org/x/time/rate` - Rate limiting

## Installation

1. Clone the repository:
```bash
git clone https://github.com/SUSHANT-0210/URL_Shortener.git
cd URL_Shortener
```

2. Install dependencies:
```bash
go mod tidy
```

3. Set environment variables (optional):
```bash
export JWT_SECRET="your-super-secret-jwt-key"
```

4. Build the application:
```bash
go build -o url-shortener
```

## Usage

### Starting the Server

Run the compiled binary:
```bash
./url-shortener
```

The server will start on `http://localhost:8080` and display available endpoints.

## API Endpoints

### Public Endpoints

#### 1. Register User
**Endpoint:** `POST /register`

**Request Body:**
```json
{
  "username": "your_username",
  "password": "your_secure_password"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "message": "User registered successfully"
}
```

#### 2. Login User
**Endpoint:** `POST /login`

**Request Body:**
```json
{
  "username": "your_username", 
  "password": "your_secure_password"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "message": "Login successful"
}
```

#### 3. Redirect to Original URL
**Endpoint:** `GET /redirect/{shortURL}`

**Example:**
```
GET /redirect/a1b2c3d4
```
Automatically redirects to the original URL.

### Protected Endpoints (Require JWT Token)

#### 4. Shorten URL
**Endpoint:** `POST /shorten`

**Headers:**
```
Authorization: Bearer your_jwt_token_here
Content-Type: application/json
```

**Request Body:**
```json
{
  "url": "https://example.com/very/long/url/that/needs/shortening"
}
```

**Response:**
```json
{
  "short_url": "a1b2c3d4",
  "id": "a1b2c3d4"
}
```

#### 5. Get User's URLs
**Endpoint:** `GET /urls`

**Headers:**
```
Authorization: Bearer your_jwt_token_here
```

**Response:**
```json
[
  {
    "id": "a1b2c3d4",
    "original_url": "https://example.com/very/long/url",
    "short_url": "a1b2c3d4", 
    "created_at": "2024-01-15T10:30:00Z",
    "user_id": "user123abc"
  }
]
```

## Using with Postman

### Step 1: Register a User
1. **Method:** POST
2. **URL:** `http://localhost:8080/register`
3. **Headers:** 
   - `Content-Type: application/json`
4. **Body (raw JSON):**
```json
{
  "username": "testuser",
  "password": "securepassword123"
}
```
5. **Save the token** from the response for subsequent requests

### Step 2: Login (Alternative to Registration)
1. **Method:** POST
2. **URL:** `http://localhost:8080/login`
3. **Headers:**
   - `Content-Type: application/json`
4. **Body (raw JSON):**
```json
{
  "username": "testuser",
  "password": "securepassword123"
}
```

### Step 3: Shorten a URL
1. **Method:** POST
2. **URL:** `http://localhost:8080/shorten`
3. **Headers:**
   - `Content-Type: application/json`
   - `Authorization: Bearer YOUR_JWT_TOKEN_HERE`
4. **Body (raw JSON):**
```json
{
  "url": "https://www.google.com/search?q=golang+url+shortener"
}
```

### Step 4: Test Redirection
1. **Method:** GET
2. **URL:** `http://localhost:8080/redirect/SHORT_URL_ID`
   - Replace `SHORT_URL_ID` with the short_url from Step 3 response
3. **No headers or body needed**
4. **Expected:** Browser/Postman will redirect to original URL

### Step 5: View Your URLs
1. **Method:** GET
2. **URL:** `http://localhost:8080/urls`
3. **Headers:**
   - `Authorization: Bearer YOUR_JWT_TOKEN_HERE`
4. **No body needed**

## Database Schema

The service uses two main tables:

### Users Table
```sql
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    username TEXT UNIQUE,
    password TEXT
);
```

### URLs Table
```sql
CREATE TABLE urls (
    id TEXT PRIMARY KEY,
    original_url TEXT,
    short_url TEXT,
    created_at DATETIME,
    user_id TEXT,
    FOREIGN KEY (user_id) REFERENCES users (id)
);
```

## How It Works

### 1. User Authentication
- Users register with username/password
- Passwords are hashed using bcrypt (cost factor: 14)
- JWT tokens issued with 24-hour expiration
- User IDs generated using SHA-256 hash of username (first 16 characters)

### 2. URL Shortening Process
- SHA-256 hash computed for the original URL
- First 8 characters used as the short URL identifier
- URL stored with user association in database
- Duplicate URLs for same user are handled gracefully

### 3. Security Features
- JWT-based authentication for protected endpoints
- Password hashing using bcrypt
- Rate limiting: 1 request/second with burst capacity of 5
- User isolation: Users can only access their own URLs

### 4. Request Flow
- Public endpoints: registration, login, redirection
- Protected endpoints require valid JWT token in Authorization header
- Rate limiting applied to all endpoints
- Comprehensive error handling with appropriate HTTP status codes

## Configuration

### Environment Variables
- `JWT_SECRET`: Custom JWT signing key (default: "!@#$%^&*()_+")

### Rate Limiting
- **Rate:** 1 request per second
- **Burst:** 5 requests
- **Scope:** Global (applied to all endpoints)

## Error Handling

The service provides comprehensive error responses:

- **400 Bad Request:** Invalid input data or missing required fields
- **401 Unauthorized:** Invalid credentials or missing/invalid JWT token  
- **404 Not Found:** Short URL doesn't exist in database
- **405 Method Not Allowed:** Wrong HTTP method used
- **429 Too Many Requests:** Rate limit exceeded
- **500 Internal Server Error:** Database or system errors

## Security Considerations

- Passwords are never stored in plain text (bcrypt hashed)
- JWT tokens expire after 24 hours
- Rate limiting prevents abuse
- User data isolation (users can only access their own URLs)
- SQL injection prevention through prepared statements
- Environment variable support for sensitive configuration

## Future Enhancements

- Custom expiration dates for shortened URLs
- URL analytics and click tracking  
- Custom short URL aliases
- Bulk URL shortening
- API key authentication for enterprise users
- URL validation and malware checking
- Database migration to PostgreSQL for production scalability
- Docker containerization
- Comprehensive logging and monitoring
- URL categories and tagging system
