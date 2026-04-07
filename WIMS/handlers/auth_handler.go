package handlers

import (
	"html/template"
	"net/http"
	"wims/models"
)

var AuthTmpl = template.Must(template.ParseGlob("templates/*.html"))

// GetSession читает токен из cookie, возвращает (username, role, displayName)
func GetSession(r *http.Request) (string, string, string) {
	cookie, err := r.Cookie("session")
	if err != nil || cookie.Value == "" {
		return "", "", ""
	}

	sess := models.GetSession(cookie.Value)
	if sess == nil {
		return "", "", ""
	}

	user := models.GetUserByUsername(sess.Username)
	if user == nil {
		return "", "", ""
	}

	display := user.Username
	if user.FirstName != "" {
		display = user.FirstName
	}

	return user.Username, user.Role, display
}

// GetSessionToken возвращает токен из cookie (нужен для logout)
func GetSessionToken(r *http.Request) string {
	cookie, err := r.Cookie("session")
	if err != nil {
		return ""
	}
	return cookie.Value
}

func SetSession(w http.ResponseWriter, r *http.Request, username string) error {
	ua := r.Header.Get("User-Agent")
	ip := r.RemoteAddr

	token, err := models.CreateSession(username, ua, ip)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

func ClearSession(w http.ResponseWriter, r *http.Request) {
	token := GetSessionToken(r)
	if token != "" {
		models.DeleteSession(token)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
}

func LoginPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		u := r.FormValue("username")
		p := r.FormValue("password")

		ok, _ := models.CheckUser(u, p)
		if ok {
			if err := SetSession(w, r, u); err != nil {
				http.Error(w, "Ошибка сессии", http.StatusInternalServerError)
				return
			}
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

func RegisterPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		err := models.CreateUser(
			r.FormValue("username"),
			r.FormValue("password"),
			"user", "", "", "", "", "",
		)
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

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	ClearSession(w, r)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
