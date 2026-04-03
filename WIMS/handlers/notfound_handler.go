package handlers

import (
	"html/template"
	"net/http"
)

var notfoundTmpl = template.Must(template.ParseFiles("templates/404.html"))

// 404 страница
func NotFoundPage(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	err := notfoundTmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
	}
}
