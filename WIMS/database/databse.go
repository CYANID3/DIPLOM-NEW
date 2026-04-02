package database

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite" // <- чистый Go драйвер, CGO не нужен
)

var DB *sql.DB

func InitDB() {
	var err error

	DB, err = sql.Open("sqlite", "wims.db") // <- вместо "sqlite3" используем "sqlite"
	if err != nil {
		log.Fatal("Ошибка при открытии базы:", err)
	}

	createTables()
}

func createTables() {
	userTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE,
		password TEXT,
		role TEXT DEFAULT 'user',
		first_name TEXT,
		last_name TEXT,
		position TEXT,
		email TEXT
	);`

	productTable := `
	CREATE TABLE IF NOT EXISTS products (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		price REAL,
		quantity INTEGER
	);`

	historyTable := `
	CREATE TABLE IF NOT EXISTS history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		action TEXT,
		username TEXT,
		target TEXT,
		quantity INTEGER,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	salesTable := `
	CREATE TABLE IF NOT EXISTS sales (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		product_id INTEGER,
		quantity INTEGER,
		username TEXT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	for _, query := range []string{userTable, productTable, historyTable, salesTable} {
		_, err := DB.Exec(query)
		if err != nil {
			log.Fatal(err)
		}
	}
}
