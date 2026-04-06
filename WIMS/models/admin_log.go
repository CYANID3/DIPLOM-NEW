package models

import (
	"database/sql"
	"strings"
	"wims/database"
)

type AdminLogItem struct {
	ID           int
	Admin        string
	AdminDisplay string
	Action       string
	Target       string
	Detail       string
	Timestamp    string
	TimestampRaw string
}

// WriteAdminLog записывает действие администратора
func WriteAdminLog(admin, action, target, detail string) {
	database.DB.Exec(
		`INSERT INTO admin_log(admin, action, target, detail) VALUES(?, ?, ?, ?)`,
		admin, action, target, detail,
	)
}

// GetAdminLog возвращает журнал административных действий
func GetAdminLog() ([]AdminLogItem, error) {
	rows, err := database.DB.Query(`
		SELECT
			l.id, l.admin,
			u.first_name, u.last_name,
			l.action, l.target, l.detail, l.timestamp
		FROM admin_log l
		LEFT JOIN users u ON l.admin = u.username
		ORDER BY l.id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []AdminLogItem
	for rows.Next() {
		var item AdminLogItem
		var first, last sql.NullString
		var ts sql.NullString

		if err := rows.Scan(
			&item.ID, &item.Admin,
			&first, &last,
			&item.Action, &item.Target, &item.Detail, &ts,
		); err != nil {
			return nil, err
		}

		firstName := strings.TrimSpace(first.String)
		lastName  := strings.TrimSpace(last.String)
		fullName  := strings.TrimSpace(firstName + " " + lastName)
		if fullName != "" {
			item.AdminDisplay = fullName
		} else {
			item.AdminDisplay = item.Admin
		}

		if ts.Valid {
			item.Timestamp, item.TimestampRaw = formatTimestamp(ts.String)
		}

		result = append(result, item)
	}
	return result, rows.Err()
}
