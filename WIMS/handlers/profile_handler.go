package handlers

import (
	"html/template"
	"net/http"
	"net/url"
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

	errorMsg := r.URL.Query().Get("error")
	successMsg := r.URL.Query().Get("success")

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
			http.Redirect(w, r, "/profile?error="+url.QueryEscape("Ошибка обновления профиля"), http.StatusSeeOther)
			return
		}

		// смена пароля
		oldPass := r.FormValue("old_password")
		pass1 := r.FormValue("password1")
		pass2 := r.FormValue("password2")

		if oldPass != "" || pass1 != "" || pass2 != "" {

			ok, _ := models.CheckPassword(username, oldPass)

			if !ok {
				http.Redirect(w, r, "/profile?error="+url.QueryEscape("Неверный текущий пароль"), http.StatusSeeOther)
				return
			}

			if pass1 != pass2 {
				http.Redirect(w, r, "/profile?error="+url.QueryEscape("Пароли не совпадают"), http.StatusSeeOther)
				return
			}

			if len(pass1) < 4 {
				http.Redirect(w, r, "/profile?error="+url.QueryEscape("Пароль слишком короткий"), http.StatusSeeOther)
				return
			}

			err := models.UpdatePassword(username, pass1)
			if err != nil {
				http.Redirect(w, r, "/profile?error="+url.QueryEscape("Ошибка смены пароля"), http.StatusSeeOther)
				return
			}
		}

		http.Redirect(w, r, "/profile?success="+url.QueryEscape("Сохранено"), http.StatusSeeOther)
		return
	}

	data := map[string]interface{}{
		"Username": display,
		"Role":     role,
		"User":     user,
		"Error":    errorMsg,
		"Success":  successMsg,
	}

	err := profileTmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
	}
}
