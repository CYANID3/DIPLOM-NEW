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
			http.Redirect(w, r, editURL(target, "Ошибка обновления профиля", ""), http.StatusSeeOther)
			return
		}

		// роль
		newRole := r.FormValue("role")

		// нельзя снять с себя права
		if target == username && newRole != "admin" {
			http.Redirect(w, r, editURL(target, "Нельзя снять с себя права администратора", ""), http.StatusSeeOther)
			return
		}

		// повышение до admin — требует подтверждения паролем
		if newRole == "admin" && user.Role != "admin" {
			confirmPwd := r.FormValue("confirm_admin_password")
			ok, _ := models.CheckPassword(username, confirmPwd)
			if !ok {
				http.Redirect(w, r, editURL(target, "Неверный пароль для подтверждения повышения до администратора", ""), http.StatusSeeOther)
				return
			}
		}

		_, err = database.DB.Exec(
			"UPDATE users SET role=? WHERE username=?",
			newRole, target,
		)
		if err != nil {
			http.Redirect(w, r, editURL(target, "Ошибка обновления роли", ""), http.StatusSeeOther)
			return
		}

		// смена пароля
		pass1 := r.FormValue("password1")
		pass2 := r.FormValue("password2")

		if pass1 != "" || pass2 != "" {
			if pass1 != pass2 {
				http.Redirect(w, r, editURL(target, "Пароли не совпадают", "password"), http.StatusSeeOther)
				return
			}
			if len(pass1) < 4 {
				http.Redirect(w, r, editURL(target, "Пароль слишком короткий (минимум 4 символа)", ""), http.StatusSeeOther)
				return
			}
			err := models.UpdatePassword(target, pass1)
			if err != nil {
				http.Redirect(w, r, editURL(target, "Ошибка смены пароля", ""), http.StatusSeeOther)
				return
			}
		}

		http.Redirect(w, r, editURL(target, "", "Изменения сохранены"), http.StatusSeeOther)
		return
	}

	errorMsg := r.URL.Query().Get("error")
	successMsg := r.URL.Query().Get("success")

	data := map[string]interface{}{
		"Username":      display,
		"Role":          role,
		"User":          user,
		"Error":         errorMsg,
		"Success":       successMsg,
		"PasswordError": r.URL.Query().Get("pe") == "1",
	}

	adminEditTmpl.Execute(w, data)
}

func editURL(target, errMsg, successMsg string) string {
	base := "/admin/edit?username=" + target
	if errMsg != "" {
		return base + "&error=" + url.QueryEscape(errMsg)
	}
	if successMsg != "" {
		return base + "&success=" + url.QueryEscape(successMsg)
	}
	return base
}
