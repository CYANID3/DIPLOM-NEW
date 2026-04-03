package handlers

import (
	"html/template"
	"net/http"
	"time"
	"wims/models"
)

var AuthTmpl = template.Must(template.ParseGlob("templates/*.html"))

// Получение сессии
func GetSession(r *http.Request) (string, string) {
	cookie, err := r.Cookie("session")
	if err != nil {
		return "", ""
	}

	username := cookie.Value
	user := models.GetUserByUsername(username)
	if user == nil {
		return "", ""
	}

	return username, user.Role
}

// Установка сессии
func SetSession(w http.ResponseWriter, username string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    username,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Path:     "/",
	})
}

// Очистка сессии
func ClearSession(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		Path:     "/",
	})
}

// LoginPage
func LoginPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		ok, _ := models.CheckUser(username, password)
		if ok {
			SetSession(w, username)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		AuthTmpl.ExecuteTemplate(w, "login.html", map[string]string{
			"Error": "Неверный логин или пароль",
		})
		return
	}

	AuthTmpl.ExecuteTemplate(w, "login.html", nil)
}

// RegisterPage
func RegisterPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		if username == "" || password == "" {
			AuthTmpl.ExecuteTemplate(w, "register.html", map[string]string{
				"Error": "Заполните все поля",
			})
			return
		}

		err := models.CreateUser(username, password, "user", "", "", "", "", "")
		if err != nil {
			AuthTmpl.ExecuteTemplate(w, "register.html", map[string]string{
				"Error": "Пользователь уже существует",
			})
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	AuthTmpl.ExecuteTemplate(w, "register.html", nil)
}

// Logout
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	ClearSession(w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
