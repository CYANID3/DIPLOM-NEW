package models

import (
	"database/sql"
	"errors"
	"strings"
	"wims/database"
)

type RegradeItem struct {
	ID          int
	Username    string
	UserDisplay string
	FromID      int
	FromName    string
	FromQty     int
	ToID        int
	ToName      string
	ToQty       int
	Note        string
	Timestamp   string
	TimestampRaw string
}

// CreateRegrade — пересорт: списать fromQty у fromID, оприходовать toQty у toID
func CreateRegrade(username string, fromID, fromQty, toID, toQty int, note string) error {
	if fromQty <= 0 || toQty <= 0 {
		return errors.New("Количество должно быть больше нуля")
	}
	if fromID == toID {
		return errors.New("Товары не могут совпадать")
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// получаем данные товара-источника
	var fromName, fromBarcode string
	var fromStock int
	var fromPrice float64
	err = tx.QueryRow(
		`SELECT name, barcode, quantity, price FROM products WHERE id = ?`, fromID,
	).Scan(&fromName, &fromBarcode, &fromStock, &fromPrice)
	if err != nil {
		return errors.New("Товар-источник не найден")
	}
	if fromStock < fromQty {
		return errors.New("Недостаточно товара для списания: " + fromName)
	}

	// получаем данные товара-назначения
	var toName, toBarcode string
	var toPrice float64
	err = tx.QueryRow(
		`SELECT name, barcode, price FROM products WHERE id = ?`, toID,
	).Scan(&toName, &toBarcode, &toPrice)
	if err != nil {
		return errors.New("Товар-назначение не найден")
	}

	// списание
	if _, err = tx.Exec(
		`UPDATE products SET quantity = quantity - ? WHERE id = ?`, fromQty, fromID,
	); err != nil {
		return err
	}

	// оприходование
	if _, err = tx.Exec(
		`UPDATE products SET quantity = quantity + ? WHERE id = ?`, toQty, toID,
	); err != nil {
		return err
	}

	// запись пересорта
	if _, err = tx.Exec(
		`INSERT INTO regradings(username, from_id, from_name, from_qty, to_id, to_name, to_qty, note)
		 VALUES(?, ?, ?, ?, ?, ?, ?, ?)`,
		username, fromID, fromName, fromQty, toID, toName, toQty, note,
	); err != nil {
		return err
	}

	// история: списание
	if _, err = tx.Exec(
		`INSERT INTO history(action, username, target, barcode, quantity, price)
		 VALUES(?, ?, ?, ?, ?, ?)`,
		"regrade_out", username, fromName, fromBarcode, fromQty, fromPrice,
	); err != nil {
		return err
	}

	// история: оприходование
	if _, err = tx.Exec(
		`INSERT INTO history(action, username, target, barcode, quantity, price)
		 VALUES(?, ?, ?, ?, ?, ?)`,
		"regrade_in", username, toName, toBarcode, toQty, toPrice,
	); err != nil {
		return err
	}

	return tx.Commit()
}

// GetRegradings — список всех пересортов
func GetRegradings() ([]RegradeItem, error) {
	rows, err := database.DB.Query(`
		SELECT
			r.id, r.username,
			u.first_name, u.last_name,
			r.from_id, r.from_name, r.from_qty,
			r.to_id, r.to_name, r.to_qty,
			r.note, r.timestamp
		FROM regradings r
		LEFT JOIN users u ON r.username = u.username
		ORDER BY r.id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []RegradeItem
	for rows.Next() {
		var item RegradeItem
		var first, last sql.NullString
		var ts sql.NullString
		if err := rows.Scan(
			&item.ID, &item.Username,
			&first, &last,
			&item.FromID, &item.FromName, &item.FromQty,
			&item.ToID, &item.ToName, &item.ToQty,
			&item.Note, &ts,
		); err != nil {
			return nil, err
		}
		firstName := strings.TrimSpace(first.String)
		lastName  := strings.TrimSpace(last.String)
		full      := strings.TrimSpace(firstName + " " + lastName)
		if full != "" {
			item.UserDisplay = full
		} else {
			item.UserDisplay = item.Username
		}
		if ts.Valid {
			item.Timestamp, item.TimestampRaw = formatTimestamp(ts.String)
		}
		result = append(result, item)
	}
	return result, rows.Err()
}
