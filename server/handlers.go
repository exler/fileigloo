package server

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/exler/fileigloo/random"
	"github.com/exler/fileigloo/storage"
	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"
)

func generateFileId() string {
	return random.String(12)
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/file", http.StatusTemporaryRedirect)
}

func (s *Server) fileHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "file", map[string]interface{}{
		"maxUploadSize": s.maxUploadSize,
		"currentPage":   "file",
	})
}

func (s *Server) pasteHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "paste", map[string]interface{}{
		"maxUploadSize": s.maxUploadSize,
		"currentPage":   "paste",
	})
}

func (s *Server) apiHandler(w http.ResponseWriter, r *http.Request) {
	// Build the base URL using the same logic as BuildURL but without path fragments
	var scheme string
	if r.TLS != nil {
		scheme = "https"
	} else if headerScheme := r.Header.Get("X-Forwarded-Proto"); headerScheme != "" {
		scheme = headerScheme
	} else {
		scheme = "http"
	}

	baseURL := scheme + "://" + r.Host

	renderTemplate(w, "api", map[string]interface{}{
		"currentPage": "api",
		"baseURL":     baseURL,
	})
}

func (s *Server) loginGETHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "login", map[string]interface{}{})
}

func (s *Server) loginPOSTHandler(w http.ResponseWriter, r *http.Request) {
	password := r.FormValue("site-password")
	if bcrypt.CompareHashAndPassword([]byte(s.sitePasswordHash), []byte(password)) != nil {
		renderTemplate(w, "login", map[string]interface{}{
			"wrongPassword": true,
		})
		return
	}

	cookie := http.Cookie{
		Name:     "site_password",
		Value:    s.sitePasswordHash,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   r.TLS != nil,
	}

	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) formHandler(w http.ResponseWriter, r *http.Request) {
	if formType := r.FormValue("form-type"); formType == "file-upload" {
		s.fileUploadHandler(w, r)
	} else if formType == "pastebin" {
		s.pastebinHandler(w, r)
	} else {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
}

const defaultMaxMemory = 32 << 20 // 32 MB

func (s *Server) fileUploadHandler(w http.ResponseWriter, r *http.Request) {
	if !ValidateContentType(r.Header) {
		http.Error(w, "Request Content-Type must be 'multipart/form-data'", http.StatusBadRequest)
		return
	}

	s.logger.Debug(fmt.Sprintf("File upload request [client_ip=%s]", r.RemoteAddr))

	var file multipart.File
	var fileHeader *multipart.FileHeader
	var err error

	// Parse the multipart form
	if err = r.ParseMultipartForm(defaultMaxMemory); err != nil {
		s.logger.Error(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// Get the file from the form
	if file, fileHeader, err = r.FormFile("file"); err != nil {
		s.logger.Error(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	fileName := SanitizeFilename(fileHeader.Filename)
	contentType := fileHeader.Header.Get("Content-Type")
	contentLength := fileHeader.Size

	if s.maxUploadSize > 0 && contentLength > s.maxUploadSize {
		http.Error(w, fmt.Sprintf("File is too big! Max upload size: %dMB", s.maxUploadSize/(1024*1024)), http.StatusRequestEntityTooLarge)
		return
	}

	var fileId string
	for {
		fileId = generateFileId()
		if _, err = s.storage.Get(r.Context(), fileId); s.storage.FileNotExists(err) {
			break
		}
	}

	metadata := storage.Metadata{
		Filename:      fileName,
		ContentType:   contentType,
		ContentLength: strconv.FormatInt(contentLength, 10),
	}
	if err = s.storage.Put(r.Context(), fileId, file, metadata); err != nil {
		s.logger.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var fileUrl *url.URL
	if ShowInline(contentType) {
		fileUrl = BuildURL(r, "view", fileId)
	} else {
		fileUrl = BuildURL(r, "download", fileId)
	}

	s.logger.Info(fmt.Sprintf("New file uploaded [url=%s]", fileUrl))

	renderTemplate(w, "file", map[string]interface{}{
		"fileUrl":       fileUrl,
		"maxUploadSize": s.maxUploadSize,
		"currentPage":   "file",
	})
}

func (s *Server) pastebinHandler(w http.ResponseWriter, r *http.Request) {
	var pasteContent string
	if pasteContent = r.FormValue("paste"); pasteContent == "" {
		http.Error(w, "Paste is empty", http.StatusBadRequest)
		return
	}

	buf := []byte(pasteContent)

	file := bytes.NewReader(buf)
	fileName := "Paste"
	contentType := "text/plain"
	contentLength := int64(len(buf))

	if s.maxUploadSize > 0 && contentLength > s.maxUploadSize {
		http.Error(w, fmt.Sprintf("File is too big! Max upload size: %dMB", s.maxUploadSize/(1024*1024)), http.StatusRequestEntityTooLarge)
		return
	}

	var fileId string
	for {
		fileId = generateFileId()
		if _, err := s.storage.Get(r.Context(), fileId); s.storage.FileNotExists(err) {
			break
		}
	}

	metadata := storage.Metadata{
		Filename:      fileName,
		ContentType:   contentType,
		ContentLength: strconv.FormatInt(contentLength, 10),
	}
	if err := s.storage.Put(r.Context(), fileId, file, metadata); err != nil {
		s.logger.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	fileUrl := BuildURL(r, "view", fileId)

	s.logger.Info(fmt.Sprintf("New file uploaded [url=%s]", fileUrl))

	renderTemplate(w, "paste", map[string]interface{}{
		"fileUrl":       fileUrl,
		"maxUploadSize": s.maxUploadSize,
		"currentPage":   "paste",
	})
}

func (s *Server) downloadHandler(w http.ResponseWriter, r *http.Request) {
	fileId := SanitizeFilename(chi.URLParam(r, "fileId"))

	reader, metadata, err := s.storage.GetWithMetadata(r.Context(), fileId)
	if s.storage.FileNotExists(err) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	} else if err != nil {
		s.logger.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer reader.Close()

	var fileDisposition string
	if chi.URLParam(r, "action") == "view" {
		fileDisposition = "inline"
		if strings.HasPrefix(metadata.ContentType, "text/") {
			metadata.ContentType = "text/plain"
		}
	} else {
		fileDisposition = "attachment"
	}

	w.Header().Set("Content-Type", metadata.ContentType)
	w.Header().Set("Content-Length", metadata.ContentLength)
	w.Header().Set("Content-Disposition", fmt.Sprintf("%s; filename=%s", fileDisposition, metadata.Filename))

	// Obtain FileSeeker
	file, err := os.CreateTemp("", "fileigloo-get-")
	if err != nil {
		s.logger.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer CleanTempFile(file)

	_, err = io.Copy(file, reader)
	if err != nil {
		s.logger.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	http.ServeContent(w, r, metadata.Filename, time.Now(), file)
}
