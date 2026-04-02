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
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	users, err := models.GetAllUsers()
	if err != nil {
		http.Error(w, "Ошибка при получении пользователей", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Users":    users,
		"Username": username,
	}
	adminTmpl.ExecuteTemplate(w, "admin.html", data)
}

// CreateUserHandler - создание нового пользователя
func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	username, role := GetSession(r)
	if username == "" || role != "admin" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
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

		if userRole == "" {
			userRole = "user"
		}
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

// DeleteUserHandler - удаление пользователя
func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	username, role := GetSession(r)
	if username == "" || role != "admin" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		delUsername := r.FormValue("username")
		if delUsername == username {
			http.Error(w, "Нельзя удалить себя", http.StatusForbidden)
			return
		}
		err := models.DeleteUser(delUsername)
		if err != nil {
			http.Error(w, "Ошибка при удалении пользователя", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	}
}

// EditUserHandler - редактирование пользователя
func EditUserHandler(w http.ResponseWriter, r *http.Request) {
	username, role := GetSession(r)
	if username == "" || role != "admin" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		oldUsername := r.FormValue("old_username")
		firstName := r.FormValue("first_name")
		lastName := r.FormValue("last_name")
		middleName := r.FormValue("middle_name")
		position := r.FormValue("position")
		email := r.FormValue("email")
		newRole := r.FormValue("role")
		password := r.FormValue("password")

		if newRole == "" {
			newRole = "user"
		}

		// Обновление пароля, если введён
		if password != "" {
			err := models.UpdateUserPassword(oldUsername, password)
			if err != nil {
				http.Error(w, "Не удалось обновить пароль", http.StatusInternalServerError)
				return
			}
		}

		// Обновление остальных данных
		err := models.UpdateUser(oldUsername, firstName, lastName, middleName, position, email, newRole)
		if err != nil {
			http.Error(w, "Не удалось обновить данные пользователя", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	}
}
