package main

import (
	"log"
	"net/http"
	"time"
	"wims/database"
	"wims/handlers"
	"wims/models"
)

func createDefaultAdmin() {
	var count int
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		log.Printf("[ERROR] Ошибка проверки пользователей: %v", err)
		return
	}
	if count > 0 {
		return
	}
	if err := models.CreateUser("admin", "admin", "admin", "Admin", "User", "", "", ""); err != nil {
		log.Printf("[ERROR] Ошибка создания admin: %v", err)
		return
	}
	log.Println("[INFO]  Создан пользователь по умолчанию: admin / admin")
}

// loggingMiddleware — логирует каждый HTTP запрос
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: 200}
		next.ServeHTTP(rw, r)
		duration := time.Since(start)

		level := "[INFO] "
		if rw.status >= 500 {
			level = "[ERROR]"
		} else if rw.status >= 400 {
			level = "[WARN] "
		}

		log.Printf("%s %s %s %d %s",
			level,
			r.Method,
			r.URL.Path,
			rw.status,
			duration.Round(time.Millisecond),
		)
	})
}

// responseWriter — обёртка для перехвата статус кода
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime)

	log.Println("[INFO]  Инициализация базы данных...")
	database.InitDB()
	log.Println("[INFO]  База данных готова")

	createDefaultAdmin()

	mux := http.NewServeMux()

	// статика
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// ===== ОСНОВНЫЕ СТРАНИЦЫ =====
	mux.HandleFunc("/", handlers.IndexPage)
	mux.HandleFunc("/sell", handlers.SellProductHandler)
	mux.HandleFunc("/add", handlers.AddProductHandler)
	mux.HandleFunc("/delete", handlers.DeleteProductHandler)

	// транзакции
	mux.HandleFunc("/transaction", handlers.TransactionPage)
	mux.HandleFunc("/transaction/sell", handlers.SellTransactionHandler)
	mux.HandleFunc("/api/product", handlers.GetProductDataHandler)
	mux.HandleFunc("/api/product-by-name", handlers.GetProductByNameHandler)

	// история
	mux.HandleFunc("/history", handlers.HistoryPage)
	mux.HandleFunc("/history/export", handlers.ExportHistoryCSVHandler)

	// профиль
	mux.HandleFunc("/profile", handlers.ProfilePage)

	// ===== АВТОРИЗАЦИЯ =====
	mux.HandleFunc("/login", handlers.LoginPage)
	mux.HandleFunc("/register", handlers.RegisterPage)
	mux.HandleFunc("/logout", handlers.LogoutHandler)

	// ===== АДМИН-ПАНЕЛЬ =====
	mux.HandleFunc("/admin", handlers.AdminPage)
	mux.HandleFunc("/admin/create", handlers.CreateUserHandler)
	mux.HandleFunc("/admin/delete", handlers.DeleteUserHandler)
	mux.HandleFunc("/admin/edit", handlers.AdminEditUserPage)
	mux.HandleFunc("/admin/dashboard", handlers.DashboardPage)

	// управление товарами
	mux.HandleFunc("/admin/products", handlers.ProductsAdminPage)
	mux.HandleFunc("/admin/products/add", handlers.AddProductAdminHandler)
	mux.HandleFunc("/admin/products/edit", handlers.EditProductPage)
	mux.HandleFunc("/admin/products/delete", handlers.DeleteProductAdminHandler)
	mux.HandleFunc("/admin/products/restock", handlers.RestockProductHandler)
	mux.HandleFunc("/admin/products/export", handlers.ExportProductsCSVHandler)

	// настройки
	mux.HandleFunc("/admin/settings", handlers.SettingsPage)

	// сессии
	mux.HandleFunc("/admin/sessions", handlers.SessionsPage)
	mux.HandleFunc("/admin/sessions/kill", handlers.KillSessionHandler)
	mux.HandleFunc("/admin/sessions/kill-all", handlers.KillUserSessionsHandler)

	// журнал
	mux.HandleFunc("/admin/log", handlers.AdminLogPage)

	// возвраты
	mux.HandleFunc("/returns", handlers.ReturnPage)
	mux.HandleFunc("/returns/create", handlers.CreateReturnHandler)
	mux.HandleFunc("/returns/export", handlers.ExportReturnsCSVHandler)

	// пересорт
	mux.HandleFunc("/regrade", handlers.RegradePage)
	mux.HandleFunc("/regrade/create", handlers.CreateRegradeHandler)

	// инвентаризация
	mux.HandleFunc("/inventory", handlers.InventoryListPage)
	mux.HandleFunc("/inventory/create", handlers.CreateInventoryHandler)
	mux.HandleFunc("/inventory/complete", handlers.CompleteInventoryHandler)
	mux.HandleFunc("/inventory/save-complete", handlers.SaveAndCompleteInventoryHandler)
	mux.HandleFunc("/inventory/export", handlers.ExportInventoryCSVHandler)
	mux.HandleFunc("/inventory/", handlers.InventoryDocPage)

	// 404
	base := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, pattern := mux.Handler(r)
		if (pattern == "/" && r.URL.Path != "/") || pattern == "" {
			handlers.NotFoundPage(w, r)
			return
		}
		mux.ServeHTTP(w, r)
	})

	logged := loggingMiddleware(base)

	log.Println("[INFO]  Сервер запущен на http://localhost:8080")
	if err := http.ListenAndServe(":8080", logged); err != nil {
		log.Fatalf("[FATAL] Ошибка запуска сервера: %v", err)
	}
}
