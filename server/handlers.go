package server

import (
	"net/http"
	"strconv"
	"time"

	"github.com/exler/fileigloo/random"
	"github.com/gorilla/mux"
)

func generateFileId() string {
	return random.String(12)
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "index", PageBody{})
}

func (s *Server) uploadHandler(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	defer file.Close()

	var fileId string
	for {
		if fileId = generateFileId(); !s.storage.FileExists(fileId) {
			break
		}
	}

	s.storage.Put(fileId, file)
	fileUrl, _ := s.router.Get("download").URL("fileId", fileId)
	renderTemplate(w, "index", PageBody{
		Message: "File uploaded successfully",
		FileUrl: r.Host + fileUrl.String(),
	})
}

func (s *Server) downloadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileId := vars["fileId"]
	if !s.storage.FileExists(fileId) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	reader, contentLength, err := s.storage.Get(fileId)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Length", strconv.FormatInt(contentLength, 10))

	http.ServeContent(w, r, fileId, time.Now(), reader)
}
