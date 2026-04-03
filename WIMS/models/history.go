package models

import (
	"database/sql"
	"wims/database"
)

type HistoryItem struct {
	ID           int
	Action       string
	Username     string
	UserFullName string
	Target       string
	Quantity     int
	Total        float64
	Timestamp    string
}

func GetHistory() ([]HistoryItem, error) {
	rows, err := database.DB.Query(`
		SELECT 
			h.id,
			h.action,
			h.username,
			u.first_name,
			u.last_name,
			h.target,
			h.quantity,
			p.price,
			h.timestamp
		FROM history h
		LEFT JOIN users u ON h.username = u.username
		LEFT JOIN products p ON p.name = h.target
		ORDER BY h.id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []HistoryItem

	for rows.Next() {
		var h HistoryItem
		var firstName, lastName sql.NullString
		var price sql.NullFloat64

		err := rows.Scan(
			&h.ID,
			&h.Action,
			&h.Username,
			&firstName,
			&lastName,
			&h.Target,
			&h.Quantity,
			&price,
			&h.Timestamp,
		)
		if err != nil {
			return nil, err
		}

		// Формируем имя
		if firstName.Valid || lastName.Valid {
			h.UserFullName = firstName.String + " " + lastName.String
		} else {
			h.UserFullName = h.Username
		}

		// Считаем сумму
		if price.Valid {
			h.Total = price.Float64 * float64(h.Quantity)
		}

		history = append(history, h)
	}

	return history, nil
}
