package server

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/argon2"
)

func SanitizeFilename(filename string) string {
	return path.Clean(path.Base(filename))
}

func ValidateContentType(h http.Header) bool {
	contentType := h.Get("Content-Type")
	if contentType == "" {
		return false
	}

	contentTypeWithoutBoundary := strings.Split(contentType, ";")[0]
	return contentTypeWithoutBoundary == "multipart/form-data"
}

func CleanTempFile(file *os.File) {
	if file != nil {
		if err := file.Close(); err != nil {
			log.Printf("Error while trying to close temp file: %s", err.Error())
		}

		if err := os.Remove(file.Name()); err != nil {
			log.Printf("Error while trying to remove temp file: %s", err.Error())
		}
	}
}

func ShowInline(contentType string) bool {
	switch {
	case
		contentType == "text/plain",
		contentType == "application/pdf",
		strings.HasPrefix(contentType, "image/"),
		strings.HasPrefix(contentType, "audio/"),
		strings.HasPrefix(contentType, "video/"):
		return true
	default:
		return false
	}
}

func BuildURL(r *http.Request, fragments ...string) *url.URL {
	var scheme string
	if r.TLS != nil {
		scheme = "https"
	} else if header_scheme := r.Header.Get("X-Forwarded-Proto"); header_scheme != "" {
		scheme = header_scheme
	} else {
		scheme = "http"
	}

	urlpath := r.Host
	for _, fragment := range fragments {
		urlpath = path.Join(urlpath, fragment)
	}
	return &url.URL{
		Path:   urlpath,
		Scheme: scheme,
	}
}

// Argon2id parameters
const (
	argon2Time      = 1
	argon2Memory    = 64 * 1024
	argon2Threads   = 4
	argon2KeyLength = 32
	saltLength      = 16
)

// HashPassword creates an Argon2id hash of the password
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", nil
	}

	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLength)

	// Encode salt and hash as base64 and combine them
	saltB64 := base64.StdEncoding.EncodeToString(salt)
	hashB64 := base64.StdEncoding.EncodeToString(hash)

	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s", argon2Memory, argon2Time, argon2Threads, saltB64, hashB64), nil
}

// VerifyPassword verifies a password against its Argon2id hash
func VerifyPassword(password, hashedPassword string) (bool, error) {
	if hashedPassword == "" {
		return password == "", nil
	}

	parts := strings.Split(hashedPassword, "$")
	if len(parts) != 6 || parts[1] != "argon2id" || parts[2] != "v=19" {
		return false, errors.New("invalid hash format")
	}

	// Parse parameters
	params := strings.Split(parts[3], ",")
	if len(params) != 3 {
		return false, errors.New("invalid parameters format")
	}

	var memory, time, threads uint32
	if _, err := fmt.Sscanf(params[0], "m=%d", &memory); err != nil {
		return false, err
	}
	if _, err := fmt.Sscanf(params[1], "t=%d", &time); err != nil {
		return false, err
	}
	if _, err := fmt.Sscanf(params[2], "p=%d", &threads); err != nil {
		return false, err
	}

	salt, err := base64.StdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}

	expectedHash, err := base64.StdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}

	hash := argon2.IDKey([]byte(password), salt, time, memory, uint8(threads), uint32(len(expectedHash)))

	return subtle.ConstantTimeCompare(hash, expectedHash) == 1, nil
}

// ParseExpirationHours parses expiration from form value and returns hours as int
// Returns default of 24 hours (1 day) if invalid or not provided
func ParseExpirationHours(expirationStr string) int {
	if expirationStr == "" {
		return 24 // Default 1 day
	}

	hours, err := strconv.Atoi(expirationStr)
	if err != nil || hours < 1 || hours > 24 {
		return 24 // Default 1 day if invalid
	}

	return hours
}

// CalculateExpirationTime calculates expiration time from current time plus hours
func CalculateExpirationTime(hours int) string {
	if hours <= 0 {
		return "" // No expiration
	}

	expirationTime := time.Now().Add(time.Duration(hours) * time.Hour)
	return expirationTime.Format(time.RFC3339)
}
