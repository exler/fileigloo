package server

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/exler/fileigloo/random"
)

func generateFileId() string {
	return random.String(12)
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello world!"))
}

func (s *Server) uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		file, _, err := r.FormFile("file")
		if err != nil {
			panic(err)
		}
		defer file.Close()

		fileId := generateFileId()
		filePath := fmt.Sprintf("%s%s", s.uploadDirectory, fileId)

		f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
		defer f.Close()
		io.Copy(f, file)

		w.Write([]byte("Thanks - file uploaded"))
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("405 - Method Not Allowed"))
	}
}
