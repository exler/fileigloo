package server

import (
	"embed"
	"html/template"
	"net/http"
	"time"
)

//go:embed templates/*
var TemplatesFS embed.FS

var (
	templates *template.Template
	funcMap   = template.FuncMap{
		"now": time.Now,
	}
)

func init() {
	templates = template.Must(template.New("").Funcs(funcMap).ParseFS(TemplatesFS, "templates/*.html"))
}

func renderTemplate(w http.ResponseWriter, tmpl string, data any) {
	err := templates.ExecuteTemplate(w, tmpl+".html", data)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
