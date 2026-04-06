package handlers

import (
	"html/template"
	"net/http"
	"net/url"
	"wims/models"
)

var sessionsTmpl = template.Must(template.ParseFiles("templates/sessions.html", "templates/navbar.html"))

func SessionsPage(w http.ResponseWriter, r *http.Request) {
	_, role, display, ok := RequireRole(w, r, "admin")
	if !ok {
		return
	}

	sessions, err := models.GetAllSessions()
	if err != nil {
		http.Error(w, "Ошибка загрузки сессий", http.StatusInternalServerError)
		return
	}

	settings := models.GetAllSettings()

	data := map[string]interface{}{
		"Username": display,
		"Role":     role,
		"Sessions": sessions,
		"Settings": settings,
		"Error":    r.URL.Query().Get("error"),
		"Success":  r.URL.Query().Get("success"),
	}

	if err := sessionsTmpl.Execute(w, data); err != nil {
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
	}
}

// KillSessionHandler — удаляет конкретную сессию по токену
func KillSessionHandler(w http.ResponseWriter, r *http.Request) {
	admin, _, _, ok := RequireRole(w, r, "admin")
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin/sessions", http.StatusSeeOther)
		return
	}

	token := r.FormValue("token")
	if token == "" {
		http.Redirect(w, r, "/admin/sessions?error="+url.QueryEscape("Токен не указан"), http.StatusSeeOther)
		return
	}

	// нельзя убить свою собственную сессию через эту форму
	if token == GetSessionToken(r) {
		http.Redirect(w, r, "/admin/sessions?error="+url.QueryEscape("Нельзя завершить собственную сессию здесь — используйте Выход"), http.StatusSeeOther)
		return
	}

	// получаем сессию чтобы узнать username для лога
	sess := models.GetSession(token)
	target := ""
	if sess != nil {
		target = sess.Username
	}

	models.DeleteSession(token)
	models.WriteAdminLog(admin, "kill_session", target, "принудительный разлогин")
	http.Redirect(w, r, "/admin/sessions?success="+url.QueryEscape("Сессия завершена"), http.StatusSeeOther)
}

// KillUserSessionsHandler — удаляет ВСЕ сессии пользователя
func KillUserSessionsHandler(w http.ResponseWriter, r *http.Request) {
	admin, _, _, ok := RequireRole(w, r, "admin")
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin/sessions", http.StatusSeeOther)
		return
	}

	target := r.FormValue("username")
	if target == "" {
		http.Redirect(w, r, "/admin/sessions?error="+url.QueryEscape("Пользователь не указан"), http.StatusSeeOther)
		return
	}

	models.DeleteUserSessions(target)
	models.WriteAdminLog(admin, "kill_all_sessions", target, "разлогин всех сессий")
	http.Redirect(w, r, "/admin/sessions?success="+url.QueryEscape("Все сессии пользователя завершены"), http.StatusSeeOther)
}
