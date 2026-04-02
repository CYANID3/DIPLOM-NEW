package handlers

import (
	"html/template"
	"net/http"
	"wims/models"
)

var profileTmpl = template.Must(template.ParseFiles("templates/profile.html"))

// ProfilePage - отображение и редактирование профиля
func ProfilePage(w http.ResponseWriter, r *http.Request) {
	username, _ := GetSession(r)
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user := models.GetUserByUsername(username)

	if r.Method == http.MethodPost {
		firstName := r.FormValue("first_name")
		lastName := r.FormValue("last_name")
		middleName := r.FormValue("middle_name")
		position := r.FormValue("position")
		email := r.FormValue("email")

		models.UpdateProfile(username, firstName, lastName, middleName, position, email)
		// Обновляем данные пользователя
		user = models.GetUserByUsername(username)
	}

	data := map[string]interface{}{
		"Username": username,
		"User":     user,
	}

	profileTmpl.Execute(w, data)
}
