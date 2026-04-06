package handlers

import (
	"html/template"
	"net/http"
	"net/url"
	"wims/database"
	"wims/models"
)

var adminEditTmpl = template.Must(template.ParseFiles("templates/admin_edit.html", "templates/navbar.html"))

func AdminEditUserPage(w http.ResponseWriter, r *http.Request) {
	admin, role, display, ok := RequireRole(w, r, "admin")
	if !ok {
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
		if err := models.UpdateProfile(
			target,
			r.FormValue("first_name"),
			r.FormValue("last_name"),
			r.FormValue("middle_name"),
			r.FormValue("position"),
			r.FormValue("email"),
		); err != nil {
			http.Redirect(w, r, editURL(target, "Ошибка обновления профиля", ""), http.StatusSeeOther)
			return
		}

		// роль
		newRole   := r.FormValue("role")
		oldRole   := user.Role

		if target == admin && newRole != "admin" {
			http.Redirect(w, r, editURL(target, "Нельзя снять с себя права администратора", ""), http.StatusSeeOther)
			return
		}

		// повышение до admin требует пароль
		if newRole == "admin" && oldRole != "admin" {
			confirmPwd := r.FormValue("confirm_admin_password")
			ok, _ := models.CheckPassword(admin, confirmPwd)
			if !ok {
				http.Redirect(w, r, editURL(target, "Неверный пароль для подтверждения повышения", ""), http.StatusSeeOther)
				return
			}
		}

		if _, err := database.DB.Exec(
			`UPDATE users SET role=? WHERE username=?`, newRole, target,
		); err != nil {
			http.Redirect(w, r, editURL(target, "Ошибка обновления роли", ""), http.StatusSeeOther)
			return
		}

		if newRole != oldRole {
			models.WriteAdminLog(admin, "change_role", target,
				oldRole+" → "+newRole)
			// если роль снижена — разлогиниваем пользователя
			if oldRole == "admin" && newRole != "admin" {
				models.DeleteUserSessions(target)
			}
		}

		// смена пароля
		pass1 := r.FormValue("password1")
		pass2 := r.FormValue("password2")
		if pass1 != "" || pass2 != "" {
			if pass1 != pass2 {
				http.Redirect(w, r, editURL(target, "Пароли не совпадают", "")+"&pe=1", http.StatusSeeOther)
				return
			}
			if len(pass1) < 4 {
				http.Redirect(w, r, editURL(target, "Пароль слишком короткий (минимум 4 символа)", ""), http.StatusSeeOther)
				return
			}
			if err := models.UpdatePassword(target, pass1); err != nil {
				http.Redirect(w, r, editURL(target, "Ошибка смены пароля", ""), http.StatusSeeOther)
				return
			}
			models.WriteAdminLog(admin, "change_password", target, "")
		}

		http.Redirect(w, r, editURL(target, "", "Изменения сохранены"), http.StatusSeeOther)
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
