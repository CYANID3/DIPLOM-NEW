package main

import (
	"log"
	"net/http"
	"wims/database"
	"wims/handlers"
	"wims/models"
)

func main() {
	database.InitDB()

	// Создание админа
	var count int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		log.Fatal("Ошибка при подсчёте пользователей:", err)
	}

	if count == 0 {
		err := models.CreateUser(
			"admin",
			"admin123",
			"admin",
			"Админ",
			"",
			"",
			"Администратор",
			"",
		)
		if err != nil {
			log.Fatal("Не удалось создать админа:", err)
		}
		log.Println("Создан админ: admin / admin123")
	}

	mux := http.NewServeMux()

	// Роуты
	mux.HandleFunc("/", handlers.IndexPage)
	mux.HandleFunc("/add", handlers.AddProductHandler)
	mux.HandleFunc("/delete", handlers.DeleteProductHandler)
	mux.HandleFunc("/sell", handlers.SellProductHandler)

	mux.HandleFunc("/login", handlers.LoginPage)
	mux.HandleFunc("/register", handlers.RegisterPage)
	mux.HandleFunc("/logout", handlers.LogoutHandler)

	mux.HandleFunc("/profile", handlers.ProfilePage)
	mux.HandleFunc("/history", handlers.HistoryPage)

	mux.HandleFunc("/admin", handlers.AdminPage)
	mux.HandleFunc("/admin/create", handlers.CreateUserHandler)
	mux.HandleFunc("/admin/delete", handlers.DeleteUserHandler)
	mux.HandleFunc("/admin/edit", handlers.EditUserHandler)

	// Обёртка с 404
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, pattern := mux.Handler(r)
		if pattern == "" {
			handlers.NotFoundPage(w, r)
			return
		}
		mux.ServeHTTP(w, r)
	})

	log.Println("Сервер запущен на :8080")

	err = http.ListenAndServe(":8080", logMiddleware(handler))
	if err != nil {
		log.Fatal("Ошибка сервера:", err)
	}
}

// Логирование запросов
func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}
