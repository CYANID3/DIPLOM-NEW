package handlers

import (
	"html/template"
	"net/http"
	"time"
	"wims/models"
)

var AuthTmpl = template.Must(template.ParseGlob("templates/*.html"))

func GetSession(r *http.Request) (string, string, string) {
	cookie, err := r.Cookie("session")
	if err != nil {
		return "", "", ""
	}

	username := cookie.Value
	user := models.GetUserByUsername(username)

	if user == nil {
		return "", "", ""
	}

	display := username
	if user.FirstName != "" {
		display = user.FirstName
	}

	return username, user.Role, display
}

func SetSession(w http.ResponseWriter, username string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    username,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Path:     "/",
	})
}

func ClearSession(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		Path:     "/",
	})
}

func LoginPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		u := r.FormValue("username")
		p := r.FormValue("password")

		ok, user := models.CheckUser(u, p)

		if ok && user != nil {
			SetSession(w, u)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		err := AuthTmpl.ExecuteTemplate(w, "login.html", map[string]string{
			"Error": "Неверный логин или пароль",
		})
		if err != nil {
			http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
		}
		return
	}

	err := AuthTmpl.ExecuteTemplate(w, "login.html", nil)
	if err != nil {
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
	}
}

func RegisterPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		err := models.CreateUser(
			r.FormValue("username"),
			r.FormValue("password"),
			"user", "", "", "", "", "",
		)

		if err != nil {
			err := AuthTmpl.ExecuteTemplate(w, "register.html", map[string]string{
				"Error": "Ошибка регистрации",
			})
			if err != nil {
				http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
			}
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	err := AuthTmpl.ExecuteTemplate(w, "register.html", nil)
	if err != nil {
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	ClearSession(w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
