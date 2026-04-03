package database

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB() {
	var err error

	DB, err = sql.Open("sqlite", "wims.db")
	if err != nil {
		log.Fatal(err)
	}

	_, err = DB.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		log.Fatal(err)
	}

	if err := DB.Ping(); err != nil {
		log.Fatal(err)
	}

	createTables()
	migrate()
}

func createTables() {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE,
			password TEXT,
			role TEXT,
			first_name TEXT,
			last_name TEXT,
			middle_name TEXT,
			position TEXT,
			email TEXT
		);`,

		`CREATE TABLE IF NOT EXISTS products (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE,
			barcode TEXT,
			price REAL,
			quantity INTEGER
		);`,

		`CREATE TABLE IF NOT EXISTS history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			action TEXT,
			username TEXT,
			target TEXT,
			barcode TEXT,
			quantity INTEGER,
			price REAL DEFAULT 0,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
	}

	for _, q := range queries {
		_, err := DB.Exec(q)
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Println("База данных и таблицы инициализированы")
}

// migrate добавляет колонку price в history если её ещё нет
// (для существующих баз данных)
func migrate() {
	_, err := DB.Exec(`ALTER TABLE history ADD COLUMN price REAL DEFAULT 0`)
	if err != nil {
		// колонка уже есть — норм
		return
	}
	log.Println("Миграция: добавлена колонка price в history")
}
