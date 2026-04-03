package handlers

import (
	"html/template"
	"net/http"
	"wims/models"
)

var adminTmpl = template.Must(template.ParseFiles("templates/admin.html"))

func AdminPage(w http.ResponseWriter, r *http.Request) {
	username, role, display := GetSession(r)

	if username == "" || role != "admin" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	users, err := models.GetAllUsers()
	if err != nil {
		http.Error(w, "Ошибка загрузки пользователей", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Username": display,
		"Role":     role,
		"Users":    users,
	}

	err = adminTmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
	}
}

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	username, role, _ := GetSession(r)

	if username == "" || role != "admin" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	err := models.CreateUser(
		r.FormValue("username"),
		r.FormValue("password"),
		r.FormValue("role"),
		r.FormValue("first_name"),
		r.FormValue("last_name"),
		r.FormValue("middle_name"),
		r.FormValue("position"),
		"",
	)

	if err != nil {
		http.Error(w, "Ошибка создания пользователя", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	username, role, _ := GetSession(r)

	if username == "" || role != "admin" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	del := r.FormValue("username")

	if del == username {
		http.Error(w, "Нельзя удалить себя", http.StatusForbidden)
		return
	}

	err := models.DeleteUser(del)
	if err != nil {
		http.Error(w, "Ошибка удаления", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
