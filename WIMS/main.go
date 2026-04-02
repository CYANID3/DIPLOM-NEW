package main

import (
	"html/template"
	"log"
	"net/http"
	"wims/database"
	"wims/handlers"
	"wims/models"
)

func main() {
	database.InitDB()

	// Создание админа при первом запуске
	var count int
	database.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if count == 0 {
		err := models.CreateUser(
			"admin",
			"admin123",
			"admin",
			"Admin",
			"Admin",
			"",
			"Администратор",
			"",
		)
		if err != nil {
			log.Fatal("Не удалось создать админа:", err)
		}
		log.Println("Создан админ: admin / admin123")
	}

	// Основные маршруты
	http.HandleFunc("/", handlers.IndexPage)
	http.HandleFunc("/add", handlers.AddProductHandler)
	http.HandleFunc("/delete", handlers.DeleteProductHandler)
	http.HandleFunc("/sell", handlers.SellProductHandler)

	// Авторизация
	http.HandleFunc("/login", handlers.LoginPage)
	http.HandleFunc("/register", handlers.RegisterPage)
	http.HandleFunc("/logout", handlers.LogoutHandler)

	// Профиль
	http.HandleFunc("/profile", handlers.ProfilePage)
	http.HandleFunc("/profile/edit", handlers.EditProfilePage)

	// История
	http.HandleFunc("/history", handlers.HistoryPage)

	// Админка
	http.HandleFunc("/admin", handlers.AdminPage)
	http.HandleFunc("/admin/create", handlers.CreateUserHandler)
	http.HandleFunc("/admin/delete", handlers.DeleteUserHandler)
	http.HandleFunc("/admin/edit", handlers.EditUserHandler)

	// 404 обработчик
	http.HandleFunc("/404", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("templates/404.html")
		if err != nil {
			http.Error(w, "404", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNotFound)
		tmpl.Execute(w, nil)
	})

	// Обёртка для 404
	http.HandleFunc("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))).ServeHTTP)

	// Кастомный NotFound
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			handlers.IndexPage(w, r)
		default:
			w.WriteHeader(http.StatusNotFound)
			tmpl, _ := template.ParseFiles("templates/404.html")
			tmpl.Execute(w, nil)
		}
	})

	log.Println("Сервер запущен на :8080")
	http.ListenAndServe(":8080", nil)
}
