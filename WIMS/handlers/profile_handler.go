package handlers

import (
	"html/template"
	"net/http"
	"wims/database"
	"wims/models"
)

var profileTmpl = template.Must(template.ParseFiles("templates/profile.html"))

func UpdateProfile(username, firstName, lastName, middleName, position, email string) error {
	_, err := database.DB.Exec(
		`UPDATE users 
		 SET first_name=?, last_name=?, middle_name=?, position=?, email=? 
		 WHERE username=?`,
		firstName, lastName, middleName, position, email, username,
	)
	return err
}

func ProfilePage(w http.ResponseWriter, r *http.Request) {
	username, role, display := GetSession(r)

	if username == "" {
		http.Redirect(w, r, "/login", 303)
		return
	}

	user := models.GetUserByUsername(username)
	message := ""

	if r.Method == http.MethodPost {

		// обновление профиля
		models.UpdateProfile(
			username,
			r.FormValue("first_name"),
			r.FormValue("last_name"),
			r.FormValue("middle_name"),
			r.FormValue("position"),
			r.FormValue("email"),
		)

		// смена пароля
		oldPass := r.FormValue("old_password")
		pass1 := r.FormValue("password1")
		pass2 := r.FormValue("password2")

		if oldPass != "" || pass1 != "" || pass2 != "" {

			ok, _ := models.CheckPassword(username, oldPass)

			if !ok {
				message = "Неверный текущий пароль"
			} else if pass1 != pass2 {
				message = "Пароли не совпадают"
			} else if len(pass1) < 4 {
				message = "Пароль слишком короткий"
			} else {
				models.UpdatePassword(username, pass1)
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

	profileTmpl.Execute(w, data)
}
