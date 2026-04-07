package handlers

import (
	"html/template"
	"net/http"
	"net/url"
	"wims/models"
)

var profileTmpl = template.Must(template.ParseFiles("templates/profile.html", "templates/navbar.html"))

func ProfilePage(w http.ResponseWriter, r *http.Request) {
	username, role, display, ok := RequireAuth(w, r)
	if !ok {
		return
	}

	user := models.GetUserByUsername(username)

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
				http.Redirect(w, r, "/profile?error="+url.QueryEscape("Пароли не совпадают")+"&password_error=1", http.StatusSeeOther)
				return
			}

			if len(pass1) < 4 {
				http.Redirect(w, r, "/profile?error="+url.QueryEscape("Пароль слишком короткий (минимум 4 символа)"), http.StatusSeeOther)
				return
			}

			err := models.UpdatePassword(username, pass1)
			if err != nil {
				http.Redirect(w, r, "/profile?error="+url.QueryEscape("Ошибка смены пароля"), http.StatusSeeOther)
				return
			}
		}

		http.Redirect(w, r, "/profile?success="+url.QueryEscape("Изменения сохранены"), http.StatusSeeOther)
		return
	}

	settings := models.GetAllSettings()

	data := map[string]interface{}{
		"Username":      display,
		"Role":          role,
		"User":          user,
		"Settings":      settings,
		"Error":         r.URL.Query().Get("error"),
		"Success":       r.URL.Query().Get("success"),
		"PasswordError": r.URL.Query().Get("password_error") == "1",
	}

	err := profileTmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
	}
}
