package server

import (
	"encoding/json"
	"net/http"
)

func SendJSON(w http.ResponseWriter, data interface{}) {
	response, err := json.Marshal(data)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(response) //#nosec
}

func SendPlain(w http.ResponseWriter, data string) {
	if data[len(data)-1] != '\n' {
		data += "\n"
	}

	response := []byte(data)
	w.Header().Set("Content-Type", "text/plain")
	w.Write(response) //#nosec
}
