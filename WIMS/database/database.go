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
		// пользователи
		`CREATE TABLE IF NOT EXISTS users (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			username    TEXT UNIQUE,
			password    TEXT,
			role        TEXT,
			first_name  TEXT,
			last_name   TEXT,
			middle_name TEXT,
			position    TEXT,
			email       TEXT
		);`,

		// товары
		`CREATE TABLE IF NOT EXISTS products (
			id        INTEGER PRIMARY KEY AUTOINCREMENT,
			name      TEXT UNIQUE,
			barcode   TEXT,
			price     REAL,
			quantity  INTEGER,
			category  TEXT    DEFAULT '',
			min_stock INTEGER DEFAULT 0
		);`,

		// история операций с товарами
		`CREATE TABLE IF NOT EXISTS history (
			id        INTEGER PRIMARY KEY AUTOINCREMENT,
			action    TEXT,
			username  TEXT,
			target    TEXT,
			barcode   TEXT,
			quantity  INTEGER,
			price     REAL    DEFAULT 0,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,

		// журнал административных действий
		`CREATE TABLE IF NOT EXISTS admin_log (
			id        INTEGER PRIMARY KEY AUTOINCREMENT,
			admin     TEXT,
			action    TEXT,
			target    TEXT,
			detail    TEXT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,

		// активные сессии
		`CREATE TABLE IF NOT EXISTS sessions (
			token      TEXT PRIMARY KEY,
			username   TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_seen  DATETIME DEFAULT CURRENT_TIMESTAMP,
			user_agent TEXT DEFAULT '',
			ip         TEXT DEFAULT ''
		);`,

		// настройки системы
		`CREATE TABLE IF NOT EXISTS settings (
			key   TEXT PRIMARY KEY,
			value TEXT
		);`,

		// возвраты товаров
		`CREATE TABLE IF NOT EXISTS returns (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			username    TEXT,
			history_id  INTEGER,
			product_id  INTEGER,
			product_name TEXT,
			barcode     TEXT,
			quantity    INTEGER,
			price       REAL DEFAULT 0,
			note        TEXT DEFAULT '',
			timestamp   DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,

		// пересорт
		`CREATE TABLE IF NOT EXISTS regradings (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			username     TEXT,
			from_id      INTEGER,
			from_name    TEXT,
			from_qty     INTEGER,
			to_id        INTEGER,
			to_name      TEXT,
			to_qty       INTEGER,
			note         TEXT DEFAULT '',
			timestamp    DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,

		// инвентаризации (документы)
		`CREATE TABLE IF NOT EXISTS inventories (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			username    TEXT,
			status      TEXT DEFAULT 'draft',
			note        TEXT DEFAULT '',
			created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
			completed_at DATETIME
		);`,

		// строки инвентаризации
		`CREATE TABLE IF NOT EXISTS inventory_items (
			id             INTEGER PRIMARY KEY AUTOINCREMENT,
			inventory_id   INTEGER NOT NULL,
			product_id     INTEGER NOT NULL,
			product_name   TEXT,
			barcode        TEXT,
			expected_qty   INTEGER,
			actual_qty     INTEGER DEFAULT 0,
			diff           INTEGER DEFAULT 0,
			price          REAL DEFAULT 0
		);`,
	}

	for _, q := range queries {
		if _, err := DB.Exec(q); err != nil {
			log.Fatal(err)
		}
	}

	seedSettings()
	log.Println("БД инициализирована")
}

// seedSettings — вставляет дефолтные настройки если их ещё нет
func seedSettings() {
	defaults := map[string]string{
		"org_name":           "WIMS",
		"currency":           "сом",
		"sell_confirm_limit": "10",
		"low_stock_limit":    "5",
		"session_lifetime":   "24",
	}
	for k, v := range defaults {
		DB.Exec(
			`INSERT OR IGNORE INTO settings(key, value) VALUES(?, ?)`,
			k, v,
		)
	}
}

// migrate — безопасно добавляет колонки в существующие таблицы
func migrate() {
	migrations := []string{
		`ALTER TABLE history  ADD COLUMN price     REAL    DEFAULT 0`,
		`ALTER TABLE products ADD COLUMN category  TEXT    DEFAULT ''`,
		`ALTER TABLE products ADD COLUMN min_stock INTEGER DEFAULT 0`,
		`CREATE TABLE IF NOT EXISTS returns (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			username     TEXT,
			history_id   INTEGER,
			product_id   INTEGER,
			product_name TEXT,
			barcode      TEXT,
			quantity     INTEGER,
			price        REAL DEFAULT 0,
			note         TEXT DEFAULT '',
			timestamp    DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS regradings (
			id        INTEGER PRIMARY KEY AUTOINCREMENT,
			username  TEXT,
			from_id   INTEGER,
			from_name TEXT,
			from_qty  INTEGER,
			to_id     INTEGER,
			to_name   TEXT,
			to_qty    INTEGER,
			note      TEXT DEFAULT '',
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS inventories (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			username     TEXT,
			status       TEXT DEFAULT 'draft',
			note         TEXT DEFAULT '',
			created_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
			completed_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS inventory_items (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			inventory_id INTEGER NOT NULL,
			product_id   INTEGER NOT NULL,
			product_name TEXT,
			barcode      TEXT,
			expected_qty INTEGER,
			actual_qty   INTEGER DEFAULT 0,
			diff         INTEGER DEFAULT 0,
			price        REAL DEFAULT 0,
			reason       TEXT DEFAULT ''
		)`,
		`ALTER TABLE inventory_items ADD COLUMN reason TEXT DEFAULT ''`,
	}
	for _, m := range migrations {
		DB.Exec(m) // ошибка = колонка уже есть, игнорируем
	}
	log.Println("Миграции применены")
}
