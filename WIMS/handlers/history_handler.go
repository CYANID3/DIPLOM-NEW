package handlers

import (
	"html/template"
	"net/http"
	"wims/models"
)

var historyTmpl = template.Must(template.ParseFiles("templates/history.html"))

func HistoryPage(w http.ResponseWriter, r *http.Request) {
	username, role, display := GetSession(r)

	if username == "" {
		http.Redirect(w, r, "/login", 303)
		return
	}

	history, err := models.GetHistory()
	if err != nil {
		http.Error(w, "Ошибка истории", 500)
		return
	}

	data := map[string]interface{}{
		"History":  history,
		"Username": display,
		"Role":     role,
	}

	historyTmpl.Execute(w, data)
}
