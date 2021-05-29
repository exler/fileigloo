package server

import (
	"html/template"
	"net/http"
	"net/url"
)

type PageBody struct {
	Host    string
	Message string
	FileUrl *url.URL
}

var templates *template.Template

func init() {
	templates = template.Must(template.ParseGlob("templates/*.html"))
}

func renderTemplate(w http.ResponseWriter, tmpl string, body PageBody) {
	err := templates.ExecuteTemplate(w, tmpl+".html", body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
