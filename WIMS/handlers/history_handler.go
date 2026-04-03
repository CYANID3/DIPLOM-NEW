package handlers

import (
	"html/template"
	"net/http"
	"wims/models"
)

var historyTmpl = template.Must(template.ParseFiles("templates/history.html"))

func HistoryPage(w http.ResponseWriter, r *http.Request) {
	username, _ := GetSession(r)
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	history, err := models.GetHistory()
	if err != nil {
		http.Error(w, "Не удалось получить историю", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"History":  history,
		"Username": username,
	}

	err = historyTmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
	}
}
