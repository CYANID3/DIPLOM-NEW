package handlers

import (
	"html/template"
	"net/http"
	"wims/models"
)

var profileTmpl = template.Must(template.ParseFiles("templates/profile.html"))

func ProfilePage(w http.ResponseWriter, r *http.Request) {
	username, role, display := GetSession(r)

	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user := models.GetUserByUsername(username)
	message := ""

	if r.Method == http.MethodPost {

		// обновление профиля
		err := models.UpdateProfile(
			username,
			r.FormValue("first_name"),
			r.FormValue("last_name"),
			r.FormValue("middle_name"),
			r.FormValue("position"),
			r.FormValue("email"),
		)

		if err != nil {
			http.Error(w, "Ошибка обновления профиля", http.StatusInternalServerError)
			return
		}

		// смена пароля
		oldPass := r.FormValue("old_password")
		pass1 := r.FormValue("password1")
		pass2 := r.FormValue("password2")

		if oldPass != "" || pass1 != "" || pass2 != "" {

			ok, err := models.CheckPassword(username, oldPass)
			if err != nil {
				http.Error(w, "Ошибка проверки пароля", http.StatusInternalServerError)
				return
			}

			if !ok {
				message = "Неверный текущий пароль"
			} else if pass1 != pass2 {
				message = "Пароли не совпадают"
			} else if len(pass1) < 4 {
				message = "Пароль слишком короткий"
			} else {
				err := models.UpdatePassword(username, pass1)
				if err != nil {
					http.Error(w, "Ошибка смены пароля", http.StatusInternalServerError)
					return
				}
				message = "Пароль изменён"
			}
		}

		user = models.GetUserByUsername(username)
	}

	data := map[string]interface{}{
		"Username": display,
		"Role":     role,
		"User":     user,
		"Message":  message,
	}

	err := profileTmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
	}
}
