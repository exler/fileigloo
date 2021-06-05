package server

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/exler/fileigloo/random"
	"github.com/gorilla/mux"
)

// 24 Kilobits
const _24K = (1 << 3) * 24

// 4 Megabytes
const _4M = (1 << 20) * 4

func generateFileId() string {
	return random.String(12)
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "index")
}

func (s *Server) uploadHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(_24K); err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	file, filename, contentType, contentLength, err := GetUpload(r)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	defer file.Close()

	if s.maxUploadSize > 0 && contentLength > s.maxUploadSize {
		http.Error(w, http.StatusText(http.StatusRequestEntityTooLarge), http.StatusRequestEntityTooLarge)
		return
	}

	var fileId string
	for {
		fileId = generateFileId()
		if _, err = s.storage.Get(fileId); s.storage.FileNotExists(err) {
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

func (s *Server) downloadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileId := SanitizeFilename(vars["fileId"])

	fileDisposition := "attachment"
	if _, ok := r.URL.Query()["inline"]; ok {
		fileDisposition = "inline"
	}

	reader, metadata, err := s.storage.GetWithMetadata(fileId)
	if s.storage.FileNotExists(err) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer reader.Close()

	w.Header().Set("Content-Type", metadata.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(metadata.ContentLength, 10))
	w.Header().Set("Content-Disposition", fmt.Sprintf("%s; filename=%s", fileDisposition, metadata.Filename))

	// Obtain FileSeeker
	file, err := ioutil.TempFile("", "fileigloo-get-")
	if err != nil {
		log.Println("Error while trying to download: %s", err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer CleanTempFile(file)

	tmpReader := io.TeeReader(reader, file)
	for {
		b := make([]byte, _4M)
		if _, err := tmpReader.Read(b); err != io.EOF {
			break
		}

		if err != nil {
			log.Printf("Error while trying to copy to output file: %s", err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	http.ServeContent(w, r, metadata.Filename, time.Now(), file)
}
