package handlers

import (
	"html/template"
	"net/http"
	"wims/models"
)

var historyTmpl = template.Must(template.ParseFiles("templates/history.html", "templates/navbar.html"))

func HistoryPage(w http.ResponseWriter, r *http.Request) {
	username, role, display, ok := RequireAuth(w, r)
	if !ok {
		return
	}

	history, err := models.GetHistory()
	if err != nil {
		http.Error(w, "Ошибка истории", http.StatusInternalServerError)
		return
	}

	settings := models.GetAllSettings()

	data := map[string]interface{}{
		"History":     history,
		"Username":    display,
		"Role":        role,
		"CurrentUser": username,
		"Settings":    settings,
	}

	if err := historyTmpl.Execute(w, data); err != nil {
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
	}
}
