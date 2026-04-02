package handlers

import (
	"html/template"
	"net/http"
)

var notfoundTmpl = template.Must(template.ParseFiles("templates/404.html"))

// Страница 404
func NotFoundPage(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	notfoundTmpl.Execute(w, nil)
}
