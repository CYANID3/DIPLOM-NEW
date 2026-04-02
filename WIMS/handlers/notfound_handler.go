package handlers

import (
	"html/template"
	"net/http"
)

var notfoundTmpl = template.Must(template.ParseGlob("templates/*.html"))

// Страница 404
func NotFoundPage(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	notfoundTmpl.ExecuteTemplate(w, "404.html", nil)
}
