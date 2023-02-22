package server

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
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
