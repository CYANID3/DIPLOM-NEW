package models

import (
	"database/sql"
	"strings"
	"wims/database"
)

type HistoryItem struct {
	ID           int
	Action       string
	Username     string
	UserFullName string
	Target       string
	Barcode      string
	Quantity     int
	Total        float64
	Timestamp    string
}

func GetHistory() ([]HistoryItem, error) {
	rows, err := database.DB.Query(`
		SELECT 
			h.id, h.action, h.username,
			u.first_name, u.last_name,
			h.target, h.barcode,
			h.quantity, p.price, h.timestamp
		FROM history h
		LEFT JOIN users u ON h.username = u.username
		LEFT JOIN products p ON p.name = h.target
		ORDER BY h.id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []HistoryItem

	for rows.Next() {
		var h HistoryItem
		var first, last sql.NullString
		var price sql.NullFloat64

		err := rows.Scan(
			&h.ID, &h.Action, &h.Username,
			&first, &last,
			&h.Target, &h.Barcode,
			&h.Quantity, &price, &h.Timestamp,
		)
		if err != nil {
			return nil, err
		}

		// формируем имя
		if first.Valid || last.Valid {
			h.UserFullName = strings.TrimSpace(first.String + " " + last.String)
		} else {
			h.UserFullName = h.Username
		}

		// считаем сумму
		if price.Valid {
			h.Total = price.Float64 * float64(h.Quantity)
		}

		result = append(result, h)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
