package main

import (
	"log"
	"net/http"
	"wims/database"
	"wims/handlers"
	"wims/models"
)

func createDefaultAdmin() {
	var count int

	err := database.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		log.Println("Ошибка проверки пользователей:", err)
		return
	}

	if count > 0 {
		log.Println("Пользователи уже существуют, admin не создаётся")
		return
	}

	err = models.CreateUser(
		"admin", "admin", "admin",
		"Admin", "User", "", "", "",
	)
	if err != nil {
		log.Println("Ошибка создания admin:", err)
		return
	}

	log.Println("Создан пользователь admin / admin")
}

func main() {
	database.InitDB()
	createDefaultAdmin()

	mux := http.NewServeMux()

	// static
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// routes — регистрируем точные паттерны
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
	mux.HandleFunc("/admin/edit", handlers.AdminEditUserPage)

	// обёртка: всё что не совпало с точным паттерном — 404
	// "/" в ServeMux матчит всё, поэтому проверяем r.URL.Path явно
	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// получаем паттерн который выбрал mux
		_, pattern := mux.Handler(r)

		// "/" матчит всё — проверяем что путь именно "/"
		if pattern == "/" && r.URL.Path != "/" {
			handlers.NotFoundPage(w, r)
			return
		}

		// пустой паттерн — тоже 404 (на случай будущих изменений)
		if pattern == "" {
			handlers.NotFoundPage(w, r)
			return
		}

		mux.ServeHTTP(w, r)
	})

	log.Println("Server :8080")
	log.Fatal(http.ListenAndServe(":8080", finalHandler))
}
