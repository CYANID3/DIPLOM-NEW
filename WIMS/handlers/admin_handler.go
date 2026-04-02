package handlers

import (
	"html/template"
	"net/http"
	"wims/models"
)

var adminTmpl = template.Must(template.ParseGlob("templates/*.html"))

// Страница админки
func AdminPage(w http.ResponseWriter, r *http.Request) {
	username, role := GetSession(r)
	if username == "" || role != "admin" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	users, err := models.GetAllUsers()
	if err != nil {
		http.Error(w, "Не удалось получить список пользователей", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Users":    users,
		"Username": username,
	}
	adminTmpl.ExecuteTemplate(w, "admin.html", data)
}

// Создание пользователя
func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	username, role := GetSession(r)
	if username == "" || role != "admin" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == "POST" {
		firstName := r.FormValue("first_name")
		lastName := r.FormValue("last_name")
		middleName := r.FormValue("middle_name")
		position := r.FormValue("position")
		login := r.FormValue("username")
		password := r.FormValue("password")
		roleNew := r.FormValue("role")

		err := models.CreateUser(login, password, roleNew, firstName, lastName, middleName, position, "")
		if err != nil {
			http.Error(w, "Не удалось создать пользователя", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

// Удаление пользователя
func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	username, role := GetSession(r)
	if username == "" || role != "admin" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == "POST" {
		delUsername := r.FormValue("username")
		models.DeleteUser(delUsername)
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

// Редактирование пользователя
func EditUserHandler(w http.ResponseWriter, r *http.Request) {
	username, role := GetSession(r)
	if username == "" || role != "admin" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == "POST" {
		editUsername := r.FormValue("username")
		firstName := r.FormValue("first_name")
		lastName := r.FormValue("last_name")
		middleName := r.FormValue("middle_name")
		position := r.FormValue("position")
		email := r.FormValue("email")
		roleNew := r.FormValue("role")

		models.UpdateUser(editUsername, firstName, lastName, middleName, position, email, roleNew)
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
