package handlers

import (
	"html/template"
	"net/http"
	"wims/models"
)

var adminTmpl = template.Must(template.ParseFiles("templates/admin.html"))

// Админ-панель
func AdminPage(w http.ResponseWriter, r *http.Request) {
	username, role := GetSession(r)
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
		"Username": username,
		"Users":    users,
	}

	err = adminTmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
	}
}

// Создание пользователя
func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	username, role := GetSession(r)
	if username == "" || role != "admin" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	firstName := r.FormValue("first_name")
	lastName := r.FormValue("last_name")
	middleName := r.FormValue("middle_name")
	position := r.FormValue("position")
	password := r.FormValue("password")
	userRole := r.FormValue("role")
	usernameInput := r.FormValue("username")

	if usernameInput == "" || password == "" {
		http.Error(w, "Логин и пароль обязательны", http.StatusBadRequest)
		return
	}

	err := models.CreateUser(usernameInput, password, userRole, firstName, lastName, middleName, position, "")
	if err != nil {
		http.Error(w, "Не удалось создать пользователя", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

// Удаление пользователя
func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	username, role := GetSession(r)
	if username == "" || role != "admin" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	delUsername := r.FormValue("username")

	if delUsername == "" {
		http.Error(w, "Не указан пользователь", http.StatusBadRequest)
		return
	}

	if delUsername == username {
		http.Error(w, "Нельзя удалить себя", http.StatusForbidden)
		return
	}

	err := models.DeleteUser(delUsername)
	if err != nil {
		http.Error(w, "Ошибка удаления пользователя", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

// Редактирование пользователя
func EditUserHandler(w http.ResponseWriter, r *http.Request) {
	username, role := GetSession(r)
	if username == "" || role != "admin" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	oldUsername := r.FormValue("old_username")
	firstName := r.FormValue("first_name")
	lastName := r.FormValue("last_name")
	middleName := r.FormValue("middle_name")
	position := r.FormValue("position")
	newRole := r.FormValue("role")
	password := r.FormValue("password")

	if oldUsername == "" {
		http.Error(w, "Не указан пользователь", http.StatusBadRequest)
		return
	}

	// Обновление пароля
	if password != "" {
		err := models.UpdateUserPassword(oldUsername, password)
		if err != nil {
			http.Error(w, "Ошибка обновления пароля", http.StatusInternalServerError)
			return
		}
	}

	// Обновление данных
	err := models.UpdateUser(oldUsername, firstName, lastName, middleName, position, "", newRole)
	if err != nil {
		http.Error(w, "Ошибка обновления пользователя", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
