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
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		log.Println("Ошибка проверки пользователей:", err)
		return
	}
	if count > 0 {
		return
	}
	if err := models.CreateUser("admin", "admin", "admin", "Admin", "User", "", "", ""); err != nil {
		log.Println("Ошибка создания admin:", err)
		return
	}
	log.Println("Создан пользователь admin / admin")
}

func main() {
	database.InitDB()
	createDefaultAdmin()

	mux := http.NewServeMux()

	// статика
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// ===== ОСНОВНЫЕ СТРАНИЦЫ =====
	mux.HandleFunc("/", handlers.IndexPage)
	mux.HandleFunc("/sell", handlers.SellProductHandler)

	// добавление и удаление — только manager/admin
	mux.HandleFunc("/add", handlers.AddProductHandler)
	mux.HandleFunc("/delete", handlers.DeleteProductHandler)

	// транзакции (продажа с выбором товара)
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

	// дашборд (admin + manager)
	mux.HandleFunc("/admin/dashboard", handlers.DashboardPage)

	// управление товарами (admin + manager)
	mux.HandleFunc("/admin/products", handlers.ProductsAdminPage)
	mux.HandleFunc("/admin/products/add", handlers.AddProductAdminHandler)
	mux.HandleFunc("/admin/products/edit", handlers.EditProductPage)
	mux.HandleFunc("/admin/products/delete", handlers.DeleteProductAdminHandler)
	mux.HandleFunc("/admin/products/restock", handlers.RestockProductHandler)
	mux.HandleFunc("/admin/products/export", handlers.ExportProductsCSVHandler)

	// настройки (admin only)
	mux.HandleFunc("/admin/settings", handlers.SettingsPage)

	// сессии (admin only)
	mux.HandleFunc("/admin/sessions", handlers.SessionsPage)
	mux.HandleFunc("/admin/sessions/kill", handlers.KillSessionHandler)
	mux.HandleFunc("/admin/sessions/kill-all", handlers.KillUserSessionsHandler)

	// журнал действий (admin only)
	mux.HandleFunc("/admin/log", handlers.AdminLogPage)

	// возвраты (manager + admin)
	mux.HandleFunc("/returns", handlers.ReturnPage)
	mux.HandleFunc("/returns/create", handlers.CreateReturnHandler)
	mux.HandleFunc("/returns/export", handlers.ExportReturnsCSVHandler)

	// пересорт (manager + admin)
	mux.HandleFunc("/regrade", handlers.RegradePage)
	mux.HandleFunc("/regrade/create", handlers.CreateRegradeHandler)

	// инвентаризация (manager + admin)
	mux.HandleFunc("/inventory", handlers.InventoryListPage)
	mux.HandleFunc("/inventory/create", handlers.CreateInventoryHandler)
	mux.HandleFunc("/inventory/complete", handlers.CompleteInventoryHandler)
	mux.HandleFunc("/inventory/save-complete", handlers.SaveAndCompleteInventoryHandler)
	mux.HandleFunc("/inventory/export", handlers.ExportInventoryCSVHandler)
	mux.HandleFunc("/inventory/", handlers.InventoryDocPage)

	// 404
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, pattern := mux.Handler(r)
		if pattern == "/" && r.URL.Path != "/" {
			handlers.NotFoundPage(w, r)
			return
		}
		if pattern == "" {
			handlers.NotFoundPage(w, r)
			return
		}
		mux.ServeHTTP(w, r)
	})

	log.Println("Server :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
