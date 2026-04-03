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
		http.Redirect(w, r, "/", 303)
		return
	}

	users, _ := models.GetAllUsers()

	data := map[string]interface{}{
		"Username": display,
		"Role":     role,
		"Users":    users,
	}

	adminTmpl.Execute(w, data)
}

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	username, role, _ := GetSession(r)

	if username == "" || role != "admin" {
		http.Redirect(w, r, "/", 303)
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
		http.Error(w, "Ошибка создания", 500)
		return
	}

	http.Redirect(w, r, "/admin", 303)
}

func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	username, role, _ := GetSession(r)

	if username == "" || role != "admin" {
		http.Redirect(w, r, "/", 303)
		return
	}

	del := r.FormValue("username")

	if del == username {
		http.Error(w, "Нельзя удалить себя", 403)
		return
	}

	models.DeleteUser(del)
	http.Redirect(w, r, "/admin", 303)
}
