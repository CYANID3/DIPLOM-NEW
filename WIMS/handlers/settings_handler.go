package handlers

import (
	"html/template"
	"net/http"
	"net/url"
	"wims/models"
)

var settingsTmpl = template.Must(template.ParseFiles("templates/settings.html", "templates/navbar.html"))

func SettingsPage(w http.ResponseWriter, r *http.Request) {
	username, role, display, ok := RequireRole(w, r, "admin")
	if !ok {
		return
	}

	if r.Method == http.MethodPost {
		keys := []string{"org_name", "currency", "sell_confirm_limit", "low_stock_limit"}
		for _, key := range keys {
			val := r.FormValue(key)
			if val == "" {
				continue
			}
			if err := models.SetSetting(key, val); err != nil {
				http.Redirect(w, r, "/admin/settings?error="+url.QueryEscape("Ошибка сохранения"), http.StatusSeeOther)
				return
			}
		}
		models.WriteAdminLog(username, "update_settings", "settings", "")
		http.Redirect(w, r, "/admin/settings?success="+url.QueryEscape("Настройки сохранены"), http.StatusSeeOther)
		return
	}

	settings := models.GetAllSettings()

	data := map[string]interface{}{
		"Username": display,
		"Role":     role,
		"Settings": settings,
		"Error":    r.URL.Query().Get("error"),
		"Success":  r.URL.Query().Get("success"),
	}

	if err := settingsTmpl.Execute(w, data); err != nil {
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
	}
}
