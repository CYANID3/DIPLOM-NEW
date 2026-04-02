package main

import (
	"log"
	"net/http"
	"wims/database"
	"wims/handlers"
	"wims/models"
)

func main() {
	// Инициализация базы
	database.InitDB()

	// Создание админа при первом запуске, если нет ни одного пользователя
	var count int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		log.Fatal("Ошибка при подсчёте пользователей:", err)
	}

	if count == 0 {
		err := models.CreateUser(
			"admin",         // username
			"admin123",      // password
			"admin",         // role
			"Админ",         // first_name
			"",              // last_name
			"",              // middle_name
			"Администратор", // position
			"",              // email
		)
		if err != nil {
			log.Fatal("Не удалось создать админа:", err)
		}
		log.Println("Создан админ: admin / admin123")
	}

	// Маршруты
	http.HandleFunc("/", handlers.IndexPage)
	http.HandleFunc("/add", handlers.AddProductHandler)
	http.HandleFunc("/delete", handlers.DeleteProductHandler)
	http.HandleFunc("/sell", handlers.SellProductHandler)

	http.HandleFunc("/login", handlers.LoginPage)
	http.HandleFunc("/register", handlers.RegisterPage)
	http.HandleFunc("/logout", handlers.LogoutHandler)

	http.HandleFunc("/profile", handlers.ProfilePage)
	http.HandleFunc("/history", handlers.HistoryPage)
	http.HandleFunc("/admin", handlers.AdminPage)
	http.HandleFunc("/admin/create", handlers.CreateUserHandler)
	http.HandleFunc("/admin/delete", handlers.DeleteUserHandler)
	http.HandleFunc("/admin/edit", handlers.EditUserHandler)

	// Страница 404
	http.HandleFunc("/404", handlers.NotFoundPage)

	log.Println("Сервер запущен на :8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Ошибка сервера:", err)
	}
}
