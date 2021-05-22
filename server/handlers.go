package server

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/exler/fileigloo/random"
	"github.com/gorilla/mux"
)

func generateFileId() string {
	return random.String(12)
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello world!"))
}

func (s *Server) uploadHandler(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("file")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var fileId, filePath string
	for {
		fileId = generateFileId()
		filePath = fmt.Sprintf("%s%s", s.uploadDirectory, fileId)

		_, err := os.Stat(filePath)
		if os.IsNotExist(err) {
			break
		}
	}

	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
	defer f.Close()

	io.Copy(f, file)
	w.Write([]byte("Thanks - file uploaded"))
}

func (s *Server) downloadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filePath := fmt.Sprintf("%s%s", s.uploadDirectory, vars["fileId"])
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		w.Write([]byte("No such file!"))
		return
	}

	http.ServeFile(w, r, filePath)
}
