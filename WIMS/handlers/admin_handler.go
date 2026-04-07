package handlers

import (
	"html/template"
	"net/http"
	"net/url"
	"wims/models"
)

var adminTmpl = template.Must(template.ParseFiles("templates/admin.html", "templates/navbar.html"))

func AdminPage(w http.ResponseWriter, r *http.Request) {
	_, role, display, ok := RequireRole(w, r, "admin")
	if !ok {
		return
	}

	users, err := models.GetAllUsers()
	if err != nil {
		http.Error(w, "Ошибка загрузки пользователей", http.StatusInternalServerError)
		return
	}

	settings := models.GetAllSettings()

	data := map[string]interface{}{
		"Username": display,
		"Role":     role,
		"Users":    users,
		"Settings": settings,
		"Error":    r.URL.Query().Get("error"),
		"Success":  r.URL.Query().Get("success"),
	}

	if err := adminTmpl.Execute(w, data); err != nil {
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
	}
}

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	admin, _, _, ok := RequireRole(w, r, "admin")
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	newUsername := r.FormValue("username")
	if newUsername == "" {
		http.Redirect(w, r, "/admin?error="+url.QueryEscape("Логин не может быть пустым"), http.StatusSeeOther)
		return
	}

	err := models.CreateUser(
		newUsername,
		r.FormValue("password"),
		r.FormValue("role"),
		r.FormValue("first_name"),
		r.FormValue("last_name"),
		r.FormValue("middle_name"),
		r.FormValue("position"),
		r.FormValue("email"),
	)
	if err != nil {
		http.Redirect(w, r, "/admin?error="+url.QueryEscape("Пользователь уже существует или ошибка создания"), http.StatusSeeOther)
		return
	}

	models.WriteAdminLog(admin, "create_user", newUsername, "роль: "+r.FormValue("role"))
	http.Redirect(w, r, "/admin?success="+url.QueryEscape("Пользователь создан"), http.StatusSeeOther)
}

func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	admin, _, _, ok := RequireRole(w, r, "admin")
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	target := r.FormValue("username")
	if target == admin {
		http.Redirect(w, r, "/admin?error="+url.QueryEscape("Нельзя удалить самого себя"), http.StatusSeeOther)
		return
	}

	adminPassword := r.FormValue("admin_password")
	ok2, _ := models.CheckPassword(admin, adminPassword)
	if !ok2 {
		http.Redirect(w, r, "/admin?error="+url.QueryEscape("Неверный пароль"), http.StatusSeeOther)
		return
	}

	// принудительно разлогиниваем удаляемого пользователя
	models.DeleteUserSessions(target)

	if err := models.DeleteUser(target); err != nil {
		http.Redirect(w, r, "/admin?error="+url.QueryEscape("Ошибка удаления пользователя"), http.StatusSeeOther)
		return
	}

	models.WriteAdminLog(admin, "delete_user", target, "")
	http.Redirect(w, r, "/admin?success="+url.QueryEscape("Пользователь удалён"), http.StatusSeeOther)
}
