package models

import (
	"database/sql"
	"wims/database"
)

// --- Структуры ---

type SalesSummary struct {
	TotalToday  float64
	TotalWeek   float64
	TotalMonth  float64
	CountToday  int
	CountWeek   int
	CountMonth  int
}

type TopProduct struct {
	Name     string
	Quantity int
	Total    float64
}

type StaffStat struct {
	UserFullName string
	Username     string
	SellCount    int
	SellTotal    float64
}

type DayStat struct {
	Day      string
	Count    int
	Total    float64
}

type LowStockProduct struct {
	ID       int
	Name     string
	Barcode  string
	Quantity int
	MinStock int
}

// --- Запросы ---

// GetSalesSummary — суммы и количество продаж за сегодня / неделю / месяц
func GetSalesSummary() (SalesSummary, error) {
	var s SalesSummary

	row := database.DB.QueryRow(`
		SELECT
			COALESCE(SUM(CASE WHEN date(timestamp) = date('now') THEN quantity * price ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN date(timestamp) >= date('now', '-7 days') THEN quantity * price ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN date(timestamp) >= date('now', '-30 days') THEN quantity * price ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN date(timestamp) = date('now') THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN date(timestamp) >= date('now', '-7 days') THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN date(timestamp) >= date('now', '-30 days') THEN 1 ELSE 0 END), 0)
		FROM history
		WHERE action = 'sell'
	`)

	err := row.Scan(
		&s.TotalToday, &s.TotalWeek, &s.TotalMonth,
		&s.CountToday, &s.CountWeek, &s.CountMonth,
	)
	return s, err
}

// GetTopProducts — топ-10 товаров по количеству продаж за 30 дней
func GetTopProducts() ([]TopProduct, error) {
	rows, err := database.DB.Query(`
		SELECT target, SUM(quantity) as qty, SUM(quantity * price) as total
		FROM history
		WHERE action = 'sell'
		  AND date(timestamp) >= date('now', '-30 days')
		GROUP BY target
		ORDER BY qty DESC
		LIMIT 10
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []TopProduct
	for rows.Next() {
		var p TopProduct
		if err := rows.Scan(&p.Name, &p.Quantity, &p.Total); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, rows.Err()
}

// GetStaffStats — статистика продаж по сотрудникам за 30 дней
func GetStaffStats() ([]StaffStat, error) {
	rows, err := database.DB.Query(`
		SELECT
			h.username,
			u.first_name, u.last_name,
			COUNT(*) as cnt,
			COALESCE(SUM(h.quantity * h.price), 0) as total
		FROM history h
		LEFT JOIN users u ON h.username = u.username
		WHERE h.action = 'sell'
		  AND date(h.timestamp) >= date('now', '-30 days')
		GROUP BY h.username
		ORDER BY total DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []StaffStat
	for rows.Next() {
		var s StaffStat
		var first, last sql.NullString
		if err := rows.Scan(&s.Username, &first, &last, &s.SellCount, &s.SellTotal); err != nil {
			return nil, err
		}
		firstName := ""
		lastName  := ""
		if first.Valid { firstName = first.String }
		if last.Valid  { lastName  = last.String  }
		full := firstName + " " + lastName
		if len([]rune(full)) > 1 {
			s.UserFullName = full
		} else {
			s.UserFullName = s.Username
		}
		result = append(result, s)
	}
	return result, rows.Err()
}

// GetDailyStats — продажи по дням за последние 30 дней
func GetDailyStats() ([]DayStat, error) {
	rows, err := database.DB.Query(`
		SELECT
			date(timestamp) as day,
			COUNT(*) as cnt,
			COALESCE(SUM(quantity * price), 0) as total
		FROM history
		WHERE action = 'sell'
		  AND date(timestamp) >= date('now', '-30 days')
		GROUP BY day
		ORDER BY day DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []DayStat
	for rows.Next() {
		var d DayStat
		var raw string
		if err := rows.Scan(&raw, &d.Count, &d.Total); err != nil {
			return nil, err
		}
		// форматируем дату
		d.Day, _ = formatTimestamp(raw + "T00:00:00Z")
		// берём только дату без времени
		if len(d.Day) >= 10 {
			d.Day = d.Day[:10]
		}
		result = append(result, d)
	}
	return result, rows.Err()
}

// GetLowStockProducts — товары у которых остаток <= min_stock
func GetLowStockProducts() ([]LowStockProduct, error) {
	rows, err := database.DB.Query(`
		SELECT id, name, barcode, quantity, min_stock
		FROM products
		WHERE min_stock > 0 AND quantity <= min_stock
		ORDER BY quantity ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []LowStockProduct
	for rows.Next() {
		var p LowStockProduct
		if err := rows.Scan(&p.ID, &p.Name, &p.Barcode, &p.Quantity, &p.MinStock); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, rows.Err()
}

// --- Дополнительная статистика ---

type ReturnsSummary struct {
	CountMonth int
	TotalMonth float64
}

type InventorySummary struct {
	CompletedCount int
	ShortageCount  int
	SurplusCount   int
	ShortageTotal  float64
	SurplusTotal   float64
}

type RegradeSummary struct {
	CountMonth int
}

// GetReturnsSummary — возвраты за 30 дней
func GetReturnsSummary() (ReturnsSummary, error) {
	var s ReturnsSummary
	err := database.DB.QueryRow(`
		SELECT
			COALESCE(COUNT(*), 0),
			COALESCE(SUM(quantity * price), 0)
		FROM returns
		WHERE date(timestamp) >= date('now', '-30 days')
	`).Scan(&s.CountMonth, &s.TotalMonth)
	return s, err
}

// GetInventorySummary — итоги инвентаризаций: недостачи и излишки из истории
func GetInventorySummary() (InventorySummary, error) {
	var s InventorySummary

	// количество завершённых инвентаризаций
	database.DB.QueryRow(
		`SELECT COUNT(*) FROM inventories WHERE status = 'completed'`,
	).Scan(&s.CompletedCount)

	// недостачи из истории
	database.DB.QueryRow(`
		SELECT COALESCE(COUNT(*), 0), COALESCE(SUM(quantity * price), 0)
		FROM history WHERE action = 'shortage'
	`).Scan(&s.ShortageCount, &s.ShortageTotal)

	// излишки из истории
	err := database.DB.QueryRow(`
		SELECT COALESCE(COUNT(*), 0), COALESCE(SUM(quantity * price), 0)
		FROM history WHERE action = 'surplus'
	`).Scan(&s.SurplusCount, &s.SurplusTotal)

	return s, err
}

// GetRegradeSummary — пересорты за 30 дней
func GetRegradeSummary() (RegradeSummary, error) {
	var s RegradeSummary
	err := database.DB.QueryRow(`
		SELECT COALESCE(COUNT(*), 0) FROM regradings
		WHERE date(timestamp) >= date('now', '-30 days')
	`).Scan(&s.CountMonth)
	return s, err
}
