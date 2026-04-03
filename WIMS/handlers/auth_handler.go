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

	display := username
	if user.FirstName != "" {
		display = user.FirstName
	}

	return username, user.Role, display
}

func SetSession(w http.ResponseWriter, username string) {
	http.SetCookie(w, &http.Cookie{
		Name:    "session",
		Value:   username,
		Expires: time.Now().Add(24 * time.Hour),
	})
}

func ClearSession(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   "session",
		Value:  "",
		MaxAge: -1,
	})
}

func LoginPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		u := r.FormValue("username")
		p := r.FormValue("password")

		ok, _ := models.CheckUser(u, p)

		if ok {
			SetSession(w, u)
			http.Redirect(w, r, "/", 303)
			return
		}

		AuthTmpl.ExecuteTemplate(w, "login.html", map[string]string{"Error": "Ошибка"})
		return
	}

	AuthTmpl.ExecuteTemplate(w, "login.html", nil)
}

func RegisterPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		err := models.CreateUser(
			r.FormValue("username"),
			r.FormValue("password"),
			"user", "", "", "", "", "",
		)

		if err != nil {
			AuthTmpl.ExecuteTemplate(w, "register.html", map[string]string{"Error": "Ошибка"})
			return
		}

		http.Redirect(w, r, "/login", 303)
		return
	}

	AuthTmpl.ExecuteTemplate(w, "register.html", nil)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	ClearSession(w)
	http.Redirect(w, r, "/login", 303)
}
