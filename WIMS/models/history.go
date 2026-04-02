package models

import (
	"wims/database"
)

type History struct {
	ID        int
	Action    string
	Username  string
	Target    string
	Quantity  int
	Total     float64 // стоимость операции
	Timestamp string
}

// Получение всей истории операций с подсчетом Total для sell
func GetAllHistory() ([]History, error) {
	rows, err := database.DB.Query(
		"SELECT id, action, username, target, quantity, timestamp FROM history ORDER BY timestamp DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []History
	for rows.Next() {
		var h History
		rows.Scan(&h.ID, &h.Action, &h.Username, &h.Target, &h.Quantity, &h.Timestamp)

		// Если действие sell, вычисляем Total
		if h.Action == "sell" {
			var price float64
			database.DB.QueryRow("SELECT price FROM products WHERE name=?", h.Target).Scan(&price)
			h.Total = price * float64(h.Quantity)
		} else {
			h.Total = 0
		}

		history = append(history, h)
	}
	return history, nil
}
