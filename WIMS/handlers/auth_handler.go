package handlers

import (
	"html/template"
	"net/http"
	"time"
	"wims/models"
)

var authTmpl = template.Must(template.ParseGlob("templates/*.html"))

// Получение текущей сессии из куки
func GetSession(r *http.Request) (string, string) {
	cookie, err := r.Cookie("session")
	if err != nil {
		return "", ""
	}
	username := cookie.Value
	user := models.GetUserByUsername(username)
	if user.ID == 0 {
		return "", ""
	}
	return username, user.Role
}

// Вход
func LoginPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		ok, _ := models.CheckUser(username, password)
		if ok {
			http.SetCookie(w, &http.Cookie{
				Name:    "session",
				Value:   username,
				Expires: time.Now().Add(24 * time.Hour),
			})
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		authTmpl.ExecuteTemplate(w, "login.html", map[string]string{"Error": "Неверный логин или пароль"})
		return
	}

	authTmpl.ExecuteTemplate(w, "login.html", nil)
}

// Регистрация
func RegisterPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		err := models.CreateUser(username, password, "user", "", "", "", "", "")
		if err != nil {
			authTmpl.ExecuteTemplate(w, "register.html", map[string]string{"Error": "Пользователь уже существует"})
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	authTmpl.ExecuteTemplate(w, "register.html", nil)
}

// Выход
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   "session",
		Value:  "",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
