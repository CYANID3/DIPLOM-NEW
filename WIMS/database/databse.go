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
		log.Fatal("Не удалось открыть базу данных:", err)
	}

	// Проверка соединения
	if err := DB.Ping(); err != nil {
		log.Fatal("Не удалось подключиться к базе данных:", err)
	}

	// Включаем foreign keys (SQLite по умолчанию выключен)
	_, err = DB.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		log.Fatal("Не удалось включить foreign keys:", err)
	}

	createTables()
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
			price REAL,
			quantity INTEGER
		);`,

		`CREATE TABLE IF NOT EXISTS history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			action TEXT,
			username TEXT,
			target TEXT,
			quantity INTEGER,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
	}

	for _, q := range queries {
		_, err := DB.Exec(q)
		if err != nil {
			log.Fatal("Ошибка при создании таблицы:", err)
		}
	}
}
