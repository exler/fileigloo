package server

import (
	"html/template"
	"net/http"
	"time"
)

var (
	templates *template.Template
	funcMap   = template.FuncMap{
		"now": time.Now,
	}
)

func init() {
	templates = template.Must(template.New("").Funcs(funcMap).ParseGlob("templates/*.html"))
}

func renderTemplate(w http.ResponseWriter, tmpl string, data any) {
	err := templates.ExecuteTemplate(w, tmpl+".html", data)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
