package handlers

import (
	"html/template"
	"net/http"
	"time"
	"wims/models"
)

var AuthTmpl = template.Must(template.ParseGlob("templates/*.html"))

// GetSession возвращает username и роль из куки
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

// SetSession сохраняет username в куки
func SetSession(w http.ResponseWriter, username string) {
	http.SetCookie(w, &http.Cookie{
		Name:    "session",
		Value:   username,
		Expires: time.Now().Add(24 * time.Hour),
	})
}

// ClearSession удаляет куки
func ClearSession(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   "session",
		Value:  "",
		MaxAge: -1,
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
		AuthTmpl.ExecuteTemplate(w, "login.html", map[string]string{"Error": "Неверный логин или пароль"})
		return
	}
	AuthTmpl.ExecuteTemplate(w, "login.html", nil)
}

// RegisterPage
func RegisterPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")
		err := models.CreateUser(username, password, "user", "", "", "", "", "")
		if err != nil {
			AuthTmpl.ExecuteTemplate(w, "register.html", map[string]string{"Error": "Пользователь уже существует"})
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	AuthTmpl.ExecuteTemplate(w, "register.html", nil)
}

// LogoutHandler
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	ClearSession(w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
