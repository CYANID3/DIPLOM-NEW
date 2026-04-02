package handlers

import (
	"html/template"
	"net/http"
	"wims/models"
)

var historyTmpl = template.Must(template.ParseGlob("templates/*.html"))

// Страница истории действий
func HistoryPage(w http.ResponseWriter, r *http.Request) {
	username, _ := GetSession(r)
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	history, err := models.GetAllHistory()
	if err != nil {
		http.Error(w, "Не удалось получить историю", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"History":  history,
		"Username": username,
	}
	historyTmpl.ExecuteTemplate(w, "history.html", data)
}
