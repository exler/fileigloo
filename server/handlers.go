package server

import (
	"bytes"
	"fmt"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/exler/fileigloo/random"
	"github.com/gorilla/mux"
)

func generateFileId() string {
	return random.String(12)
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "index")
}

func (s *Server) uploadHandler(w http.ResponseWriter, r *http.Request) {
	file, fheader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	defer file.Close()

	filename := SanitizeFilename(fheader.Filename)
	contentType := mime.TypeByExtension(filepath.Ext(fheader.Filename))
	contentLength := fheader.Size

	if s.maxUploadSize > 0 && contentLength > s.maxUploadSize {
		http.Error(w, http.StatusText(http.StatusRequestEntityTooLarge), http.StatusRequestEntityTooLarge)
		return
	}

	var fileId string
	for {
		if fileId = generateFileId(); !s.storage.FileExists(fileId) {
			break
		}
	}

	metadata := MakeMetadata(filename, contentType, contentLength)
	if err := s.storage.Put(fileId, file, metadata); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	fileUrl, err := s.router.Get("download").URL("fileId", fileId)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	response := FileUploadedResponse{
		FileUrl: r.Host + fileUrl.String(),
	}
	sendJSON(w, response)
}

func (s *Server) pasteHandler(w http.ResponseWriter, r *http.Request) {
	var paste string
	if paste = r.FormValue("paste"); paste == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	pasteBytes := []byte(paste)
	reader := bytes.NewReader(pasteBytes)

	filename := "Paste"
	contentType := "text/plain"
	contentLength := int64(len(pasteBytes))

	if s.maxUploadSize > 0 && contentLength > s.maxUploadSize {
		http.Error(w, http.StatusText(http.StatusRequestEntityTooLarge), http.StatusRequestEntityTooLarge)
		return
	}

	var fileId string
	for {
		if fileId = generateFileId(); !s.storage.FileExists(fileId) {
			break
		}
	}

	metadata := MakeMetadata(filename, contentType, contentLength)
	if err := s.storage.Put(fileId, reader, metadata); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	fileUrl, err := s.router.Get("download").URL("fileId", fileId)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	response := FileUploadedResponse{
		FileUrl: r.Host + fileUrl.String() + "?inline",
	}
	sendJSON(w, response)
}

func (s *Server) downloadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileId := vars["fileId"]

	fileDisposition := "attachment"
	if _, ok := r.URL.Query()["inline"]; ok {
		fileDisposition = "inline"
	}

	if !s.storage.FileExists(fileId) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	reader, metadata, err := s.storage.Get(fileId)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", metadata.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(metadata.ContentLength, 10))
	w.Header().Set("Content-Disposition", fmt.Sprintf("%s; filename=%s", fileDisposition, metadata.Filename))

	http.ServeContent(w, r, metadata.Filename, time.Now(), reader)
}
