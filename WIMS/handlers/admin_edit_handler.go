package handlers

import (
	"html/template"
	"net/http"
	"net/url"
	"wims/database"
	"wims/models"
)

var adminEditTmpl = template.Must(template.ParseFiles("templates/admin_edit.html"))

func AdminEditUserPage(w http.ResponseWriter, r *http.Request) {
	username, role, display := GetSession(r)

	if username == "" || role != "admin" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	target := r.URL.Query().Get("username")
	if target == "" {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	user := models.GetUserByUsername(target)
	if user == nil {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		// обновление профиля
		err := models.UpdateProfile(
			target,
			r.FormValue("first_name"),
			r.FormValue("last_name"),
			r.FormValue("middle_name"),
			r.FormValue("position"),
			r.FormValue("email"),
		)
		if err != nil {
			http.Redirect(w, r, "/admin/edit?username="+target+"&error="+url.QueryEscape("Ошибка обновления профиля"), http.StatusSeeOther)
			return
		}

		// роль — нельзя снять с себя админа
		newRole := r.FormValue("role")
		if target == username && newRole != "admin" {
			http.Redirect(w, r, "/admin/edit?username="+target+"&error="+url.QueryEscape("Нельзя снять с себя права администратора"), http.StatusSeeOther)
			return
		}

		_, err = database.DB.Exec(
			"UPDATE users SET role=? WHERE username=?",
			newRole, target,
		)
		if err != nil {
			http.Redirect(w, r, "/admin/edit?username="+target+"&error="+url.QueryEscape("Ошибка обновления роли"), http.StatusSeeOther)
			return
		}

		// смена пароля — только если хоть одно поле заполнено
		pass1 := r.FormValue("password1")
		pass2 := r.FormValue("password2")

		if pass1 != "" || pass2 != "" {
			if pass1 != pass2 {
				http.Redirect(w, r, "/admin/edit?username="+target+"&error="+url.QueryEscape("Пароли не совпадают"), http.StatusSeeOther)
				return
			}

			if len(pass1) < 4 {
				http.Redirect(w, r, "/admin/edit?username="+target+"&error="+url.QueryEscape("Пароль слишком короткий (минимум 4 символа)"), http.StatusSeeOther)
				return
			}

			err := models.UpdatePassword(target, pass1)
			if err != nil {
				http.Redirect(w, r, "/admin/edit?username="+target+"&error="+url.QueryEscape("Ошибка смены пароля"), http.StatusSeeOther)
				return
			}
		}

		http.Redirect(w, r, "/admin/edit?username="+target+"&success="+url.QueryEscape("Изменения сохранены"), http.StatusSeeOther)
		return
	}

	// GET — читаем query params после редиректа
	errorMsg := r.URL.Query().Get("error")
	successMsg := r.URL.Query().Get("success")

	data := map[string]interface{}{
		"Username":      display,
		"Role":          role,
		"User":          user,
		"Error":         errorMsg,
		"Success":       successMsg,
		"PasswordError": errorMsg == "Пароли не совпадают",
	}

	adminEditTmpl.Execute(w, data)
}
