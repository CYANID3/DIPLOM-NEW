package handlers

import (
	"log"
	"net/http"
	"strings"
)

// RequireAuth — проверяет что пользователь залогинен
func RequireAuth(w http.ResponseWriter, r *http.Request) (string, string, string, bool) {
	username, role, display := GetSession(r)
	if username == "" {
		log.Printf("[AUTH]  Неавторизованный доступ: %s %s", r.Method, r.URL.Path)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return "", "", "", false
	}
	return username, role, display, true
}

// RequireRole — проверяет роль, при несоответствии редиректит на главную с ошибкой
func RequireRole(w http.ResponseWriter, r *http.Request, roles ...string) (string, string, string, bool) {
	username, role, display := GetSession(r)
	if username == "" {
		log.Printf("[AUTH]  Неавторизованный доступ: %s %s", r.Method, r.URL.Path)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return "", "", "", false
	}
	for _, allowed := range roles {
		if role == allowed {
			return username, role, display, true
		}
	}
	log.Printf("[WARN]  Отказ в доступе: пользователь=%s роль=%s путь=%s (требуется: %s)",
		username, role, r.URL.Path, strings.Join(roles, "|"))
	http.Redirect(w, r, "/?error=Недостаточно+прав", http.StatusSeeOther)
	return "", "", "", false
}
