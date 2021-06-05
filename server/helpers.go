package server

import (
	"log"
	"os"
	"path"
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

func GetUpload(r *http.Request) (file *io.File, filename, contentType string, contentLength int64, err error) {
	var fheader *multipart.FileHeader
	var text string
	if file, fheader, err = r.FormFile("file"); err != nil {
		filename = SanitizeFilename(fheader.Filename)
		contentType = mime.TypeByExtension(filepath.Ext(fheader.Filename))
		contentLength = fheader.Size
	} else if text = r.FormValue("text"); text != "" {
		textBytes := []byte(text)
		file = bytes.NewReader(textBytes)

		filename = "Paste"
		contentType = "text/plain"
		contentLength = int64(len(textBytes))
	}

	return
}
