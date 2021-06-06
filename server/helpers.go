package server

import (
	"bytes"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
)

func SanitizeFilename(filename string) string {
	return path.Clean(path.Base(filename))
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

func GetDownloadURL(r *http.Request, fileUrl *url.URL) string {
	fileUrl.Host = r.Host
	if r.TLS != nil {
		fileUrl.Scheme = "https"
	} else {
		fileUrl.Scheme = "http"
	}

	return fileUrl.String()
}

func GetUpload(r *http.Request) (file io.Reader, filename, contentType string, contentLength int64, err error) {
	var fheader *multipart.FileHeader
	if file, fheader, err = r.FormFile("file"); err == nil {
		filename = SanitizeFilename(fheader.Filename)
		contentType = mime.TypeByExtension(filepath.Ext(fheader.Filename))
		contentLength = fheader.Size
	} else if text := r.FormValue("text"); text != "" {
		err = nil

		buf := []byte(text)
		file = bytes.NewReader(buf)

		filename = "Paste"
		contentType = "text/plain"
		contentLength = int64(len(buf))
	}

	return
}
