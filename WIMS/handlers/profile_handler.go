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

	data := struct {
		User models.User
		Role string
	}{
		User: *user,
		Role: role,
	}

	profileTmpl.ExecuteTemplate(w, "profile_view.html", data)
}

// Редактирование профиля
func EditProfilePage(w http.ResponseWriter, r *http.Request) {
	username, role := GetSession(r)
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user := models.GetUserByUsername(username)

	if r.Method == "POST" {
		firstName := r.FormValue("first_name")
		lastName := r.FormValue("last_name")
		middleName := r.FormValue("middle_name")
		position := r.FormValue("position")
		email := r.FormValue("email")
		newRole := r.FormValue("role")

		// Только админ может менять роль
		if role != "admin" {
			newRole = user.Role
		}

		err := models.UpdateUser(username, firstName, lastName, middleName, position, email, newRole)
		if err != nil {
			http.Error(w, "Не удалось обновить профиль", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	data := struct {
		User models.User
		Role string
	}{
		User: *user,
		Role: role,
	}

	profileTmpl.ExecuteTemplate(w, "profile_edit.html", data)
}
