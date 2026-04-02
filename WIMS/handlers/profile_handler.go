package handlers

import (
	"html/template"
	"net/http"
	"wims/models"
)

var profileTmpl = template.Must(template.ParseGlob("templates/*.html"))

// Просмотр профиля
func ProfilePage(w http.ResponseWriter, r *http.Request) {
	username, role := GetSession(r)
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user := models.GetUserByUsername(username)
	if user.Username == "" {
		http.Error(w, "Пользователь не найден", http.StatusNotFound)
		return
	}

	// Если роль не задана, присвоить user
	if role == "" {
		role = "user"
	}

	// Админ перенаправляется на админ-панель
	if role == "admin" {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	data := struct {
		User models.User
		Role string
	}{
		User: *user,
		Role: role,
	}

	profileTmpl.ExecuteTemplate(w, "profile.html", data)
}

// Редактирование профиля
func EditProfilePage(w http.ResponseWriter, r *http.Request) {
	username, _ := GetSession(r)
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user := models.GetUserByUsername(username)
	if user.Username == "" {
		http.Error(w, "Пользователь не найден", http.StatusNotFound)
		return
	}

	if r.Method == "POST" {
		firstName := r.FormValue("first_name")
		lastName := r.FormValue("last_name")
		middleName := r.FormValue("middle_name")
		position := r.FormValue("position")
		email := r.FormValue("email")

		err := models.UpdateProfile(username, firstName, lastName, middleName, position, email)
		if err != nil {
			http.Error(w, "Не удалось обновить профиль", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	data := struct {
		User models.User
	}{
		User: *user,
	}

	profileTmpl.ExecuteTemplate(w, "profile.html", data)
}
