package models

import (
	"database/sql"
	"strings"
	"time"
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
	Price        float64
	Total        float64
	Timestamp    string
	TimestampRaw string // для сортировки на клиенте (ISO)
}

// форматы которые может вернуть SQLite
var tsFormats = []string{
	"2006-01-02T15:04:05Z",
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
	"2006-01-02T15:04:05Z07:00",
}

func formatTimestamp(raw string) (display, iso string) {
	raw = strings.TrimSpace(raw)
	for _, layout := range tsFormats {
		t, err := time.Parse(layout, raw)
		if err == nil {
			return t.Format("02.01.2006 15:04"), t.UTC().Format("2006-01-02T15:04:05Z")
		}
	}
	// не распарсилось — вернуть как есть
	return raw, raw
}

func GetHistory() ([]HistoryItem, error) {
	rows, err := database.DB.Query(`
		SELECT
			h.id, h.action, h.username,
			u.first_name, u.last_name,
			h.target, h.barcode,
			h.quantity, h.price, h.timestamp
		FROM history h
		LEFT JOIN users u ON h.username = u.username
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
		var price      sql.NullFloat64
		var timestamp  sql.NullString

		err := rows.Scan(
			&h.ID, &h.Action, &h.Username,
			&first, &last,
			&h.Target, &h.Barcode,
			&h.Quantity, &price, &timestamp,
		)
		if err != nil {
			return nil, err
		}

		// имя: используем только если хотя бы одно непустое
		firstName := strings.TrimSpace(first.String)
		lastName  := strings.TrimSpace(last.String)
		fullName  := strings.TrimSpace(firstName + " " + lastName)
		if fullName != "" {
			h.UserFullName = fullName
		} else {
			h.UserFullName = h.Username
		}

		if price.Valid {
			h.Price = price.Float64
			h.Total = price.Float64 * float64(h.Quantity)
		}

		if timestamp.Valid {
			h.Timestamp, h.TimestampRaw = formatTimestamp(timestamp.String)
		}

		result = append(result, h)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
