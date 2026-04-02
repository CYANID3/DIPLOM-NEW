package handlers

import (
	"html/template"
	"net/http"
	"wims/models"
)

var historyTmpl = template.Must(template.ParseFiles("templates/history.html"))

func HistoryPage(w http.ResponseWriter, r *http.Request) {
	history, err := models.GetHistory()
	if err != nil {
		w.Write([]byte("Не удалось получить историю"))
		return
	}

	data := map[string]interface{}{
		"History":  history,
		"Username": "Текущий пользователь", // позже сюда можно подставить сессию
	}

	historyTmpl.Execute(w, data)
}
