package server

import (
	"encoding/json"
	"net/http"
)

type FileUploadedResponse struct {
	FileUrl string
}

func sendJSON(w http.ResponseWriter, data interface{}) {
	jsonResponse, err := json.Marshal(data)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}
