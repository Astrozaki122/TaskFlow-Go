package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"task-platform/internal/database"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte(getEnv("JWT_SECRET", "secret"))

const tokenExpiry = 24 * time.Hour

type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type apiResponse struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
	Token   string `json:"token,omitempty"` // useful for API testing
}

func writeJSON(w http.ResponseWriter, status int, resp apiResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
}

func Register(w http.ResponseWriter, r *http.Request) {

	var u User

	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		writeJSON(w, http.StatusBadRequest, apiResponse{Error: "invalid JSON body"})
		return
	}

	if u.Email == "" || u.Password == "" {
		writeJSON(w, http.StatusBadRequest, apiResponse{Error: "email and password required"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiResponse{Error: "password hashing failed"})
		return
	}

	_, err = database.DB.Exec(
		"INSERT INTO users (email, password_hash) VALUES ($1, $2)",
		u.Email,
		string(hash),
	)

	if err != nil {
		writeJSON(w, http.StatusBadRequest, apiResponse{Error: "user already exists"})
		return
	}

	writeJSON(w, http.StatusCreated, apiResponse{Message: "user created"})
}

func Login(w http.ResponseWriter, r *http.Request) {

	var u User

	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		writeJSON(w, http.StatusBadRequest, apiResponse{Error: "invalid JSON"})
		return
	}

	var userID int
	var storedHash string

	err := database.DB.QueryRow(
		"SELECT id, password_hash FROM users WHERE email=$1",
		u.Email,
	).Scan(&userID, &storedHash)

	if err != nil {
		writeJSON(w, http.StatusUnauthorized, apiResponse{Error: "invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(u.Password)); err != nil {
		writeJSON(w, http.StatusUnauthorized, apiResponse{Error: "invalid credentials"})
		return
	}

	tokenString, err := generateJWT(userID, u.Email)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiResponse{Error: "token generation failed"})
		return
	}

	setAuthCookie(w, tokenString)

	writeJSON(w, http.StatusOK, apiResponse{
		Message: "login successful",
		Token:   tokenString,
	})
}

// --------------------
// JWT GENERATION
// --------------------

func generateJWT(userID int, email string) (string, error) {

	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(tokenExpiry).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func setAuthCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false, // true in production (HTTPS)
		Expires:  time.Now().Add(tokenExpiry),
	})
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
