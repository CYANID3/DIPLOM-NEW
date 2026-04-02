package handlers

import (
	"html/template"
	"net/http"
	"wims/models"
)

var adminTmpl = template.Must(template.ParseFiles("templates/admin.html"))

// AdminPage - отображение админ-панели
func AdminPage(w http.ResponseWriter, r *http.Request) {
	username, role := GetSession(r)
	if username == "" || role != "admin" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	users, _ := models.GetAllUsers()
	data := map[string]interface{}{
		"Username": username,
		"Users":    users,
	}

	adminTmpl.Execute(w, data)
}

// CreateUserHandler
func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	username, role := GetSession(r)
	if username == "" || role != "admin" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		firstName := r.FormValue("first_name")
		lastName := r.FormValue("last_name")
		middleName := r.FormValue("middle_name")
		position := r.FormValue("position")
		password := r.FormValue("password")
		userRole := r.FormValue("role")
		usernameInput := r.FormValue("username")

		if password == "" {
			password = "12345" // дефолтный пароль
		}

		err := models.CreateUser(usernameInput, password, userRole, firstName, lastName, middleName, position, "")
		if err != nil {
			http.Error(w, "Не удалось создать пользователя", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	}
}

// DeleteUserHandler
func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	username, role := GetSession(r)
	if username == "" || role != "admin" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		delUsername := r.FormValue("username")
		if delUsername == username {
			http.Error(w, "Нельзя удалить себя", http.StatusForbidden)
			return
		}
		models.DeleteUser(delUsername)
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	}
}

// EditUserHandler
func EditUserHandler(w http.ResponseWriter, r *http.Request) {
	username, role := GetSession(r)
	if username == "" || role != "admin" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		oldUsername := r.FormValue("old_username")
		firstName := r.FormValue("first_name")
		lastName := r.FormValue("last_name")
		middleName := r.FormValue("middle_name")
		position := r.FormValue("position")
		newRole := r.FormValue("role")
		password := r.FormValue("password")

		if password != "" {
			models.UpdateUserPassword(oldUsername, password)
		}

		models.UpdateUser(oldUsername, firstName, lastName, middleName, position, "", newRole)
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	}
}
