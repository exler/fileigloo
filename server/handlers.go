package server

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/exler/fileigloo/random"
	"github.com/exler/fileigloo/storage"
	"github.com/go-chi/chi/v5"
)

// 128 Kilobits
const _128K = (1 << 3) * 128

func generateFileId() string {
	return random.String(12)
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "index", nil)
}

func (s *Server) uploadHandler(w http.ResponseWriter, r *http.Request) {
	if !ValidateContentType(r.Header) {
		http.Error(w, "Request Content-Type must be 'multipart/form-data'", http.StatusBadRequest)
		return
	}

	if err := r.ParseMultipartForm(_128K); err != nil {
		s.logger.Error(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	file, filename, contentType, contentLength, err := GetUpload(r)
	if err != nil {
		if err == http.ErrMissingFile {
			http.Error(w, "Request is missing `file` or `text` parameters", http.StatusBadRequest)
		} else {
			s.logger.Error(err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		}

		return
	}

	if s.maxUploadSize > 0 && contentLength > s.maxUploadSize {
		http.Error(w, fmt.Sprintf("File is too big! Max upload size: %dMB", s.maxUploadSize/(1024*1000)), http.StatusRequestEntityTooLarge)
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
		Filename:      filename,
		ContentType:   contentType,
		ContentLength: strconv.FormatInt(contentLength, 10),
	}
	if err := s.storage.Put(r.Context(), fileId, file, metadata); err != nil {
		s.logger.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var fileUrl *url.URL
	if ShowInline(contentType) {
		fileUrl = BuildURL(r, "view", fileId)
	} else {
		fileUrl = BuildURL(r, fileId)
	}

	s.logger.Info(fmt.Sprintf("New file uploaded [url=%s]", fileUrl))

	renderTemplate(w, "index", map[string]interface{}{
		"fileUrl":       fileUrl,
		"retentionTime": s.storage.RetentionTime().String(),
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
	if chi.URLParam(r, "view") != "" {
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
