package server

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/exler/fileigloo/random"
	"github.com/exler/fileigloo/storage"
	"github.com/getsentry/sentry-go"
	"github.com/gorilla/mux"
	"github.com/microcosm-cc/bluemonday"
)

// 128 Kilobits
const _128K = (1 << 3) * 128

// 4 Megabytes
const _4M = (1 << 20) * 4

func generateFileId() string {
	return random.String(12)
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "index")
}

func (s *Server) uploadHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(_128K); err != nil {
		sentry.CaptureException(err)
		log.Println(err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	file, filename, contentType, contentLength, err := GetUpload(r)
	if err != nil {
		sentry.CaptureException(err)
		log.Println(err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if s.maxUploadSize > 0 && contentLength > s.maxUploadSize {
		http.Error(w, fmt.Sprintf("File is too big! Max upload size: %dMB", s.maxUploadSize/(1024*1000)), http.StatusRequestEntityTooLarge)
		return
	}

	var fileId string
	for {
		fileId = generateFileId()
		if _, err = s.storage.Get(fileId); s.storage.FileNotExists(err) {
			break
		}
	}

	metadata := storage.MakeMetadata(filename, contentType, contentLength)
	if err := s.storage.Put(fileId, file, metadata); err != nil {
		sentry.CaptureException(err)
		log.Println(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var fileUrl *url.URL
	if ShowInline(contentType) {
		fileUrl, err = s.router.Get("view").URL("view", "view", "fileId", fileId)
	} else {
		fileUrl, err = s.router.Get("download").URL("fileId", fileId)
	}

	if err != nil {
		sentry.CaptureException(err)
		log.Println(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	response := s.GetDownloadURL(r, fileUrl)
	sentry.CaptureMessage(fmt.Sprintf("New file uploaded: %s", response))

	SendPlain(w, response)
}

func (s *Server) downloadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileId := SanitizeFilename(vars["fileId"])

	reader, metadata, err := s.storage.GetWithMetadata(fileId)
	if s.storage.FileNotExists(err) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	} else if err != nil {
		sentry.CaptureException(err)
		log.Println(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer reader.Close()

	var fileDisposition string
	if _, ok := vars["view"]; ok {
		fileDisposition = "inline"
		if strings.HasPrefix(metadata.ContentType, "text/") {
			reader = ioutil.NopCloser(bluemonday.UGCPolicy().SanitizeReader(reader))
		}
	} else {
		fileDisposition = "attachment"
	}

	w.Header().Set("Content-Type", metadata.ContentType)
	w.Header().Set("Content-Length", metadata.ContentLength)
	w.Header().Set("Content-Disposition", fmt.Sprintf("%s; filename=%s", fileDisposition, metadata.Filename))

	// Obtain FileSeeker
	file, err := ioutil.TempFile("", "fileigloo-get-")
	if err != nil {
		sentry.CaptureException(err)
		log.Println(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer CleanTempFile(file)

	tmpReader := io.TeeReader(reader, file)
	for {
		b := make([]byte, _4M)
		if _, err := tmpReader.Read(b); err == io.EOF {
			break
		}

		if err != nil {
			sentry.CaptureException(err)
			log.Println(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	http.ServeContent(w, r, metadata.Filename, time.Now(), file)
}
