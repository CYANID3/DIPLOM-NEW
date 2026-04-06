package handlers

import (
	"html/template"
	"net/http"
	"wims/models"
)

var adminLogTmpl = template.Must(template.ParseFiles("templates/admin_log.html", "templates/navbar.html"))

func AdminLogPage(w http.ResponseWriter, r *http.Request) {
	_, role, display, ok := RequireRole(w, r, "admin")
	if !ok {
		return
	}

	logs, err := models.GetAdminLog()
	if err != nil {
		http.Error(w, "Ошибка загрузки журнала", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Username": display,
		"Role":     role,
		"Logs":     logs,
	}

	if err := adminLogTmpl.Execute(w, data); err != nil {
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
	}
}
