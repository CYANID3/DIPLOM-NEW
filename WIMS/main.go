package main

import (
	"log"
	"net/http"
	"wims/database"
	"wims/handlers"
)

func main() {
	database.InitDB()

	mux := http.NewServeMux()

	// static
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// routes
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

	// обёртка для 404
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, pattern := mux.Handler(r)

		if pattern == "" {
			handlers.NotFoundPage(w, r)
			return
		}

		mux.ServeHTTP(w, r)
	})

	log.Println("Server :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
