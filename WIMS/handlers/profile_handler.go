package handlers

import (
	"html/template"
	"net/http"
	"wims/models"
)

var profileTmpl = template.Must(template.ParseFiles("templates/profile.html"))

// Профиль
func ProfilePage(w http.ResponseWriter, r *http.Request) {
	username, role := GetSession(r)
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user := models.GetUserByUsername(username)
	if user == nil {
		http.Error(w, "Пользователь не найден", http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodPost {
		firstName := r.FormValue("first_name")
		lastName := r.FormValue("last_name")
		middleName := r.FormValue("middle_name")
		position := r.FormValue("position")
		email := r.FormValue("email")

		err := models.UpdateProfile(username, firstName, lastName, middleName, position, email)
		if err != nil {
			http.Error(w, "Ошибка обновления профиля", http.StatusInternalServerError)
			return
		}

		user = models.GetUserByUsername(username)
	}

	data := map[string]interface{}{
		"Username": username,
		"User":     user,
		"Role":     role,
	}

	err := profileTmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
	}
}
