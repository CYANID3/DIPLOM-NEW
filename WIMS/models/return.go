package models

import (
	"database/sql"
	"errors"
	"strings"
	"wims/database"
)

type ReturnItem struct {
	ID          int
	Username    string
	UserDisplay string
	HistoryID   int
	ProductID   int
	ProductName string
	Barcode     string
	Quantity    int
	Price       float64
	Total       float64
	Note        string
	Timestamp   string
	TimestampRaw string
}

// SellHistoryItem — запись продажи для выбора при возврате
type SellHistoryItem struct {
	ID          int
	ProductName string
	Barcode     string
	Quantity    int
	Price       float64
	Total       float64
	Timestamp   string
	TimestampRaw string
}

// GetSellHistory — последние 100 продаж для выбора возврата
func GetSellHistory() ([]SellHistoryItem, error) {
	rows, err := database.DB.Query(`
		SELECT id, target, barcode, quantity, price, timestamp
		FROM history
		WHERE action = 'sell'
		ORDER BY id DESC
		LIMIT 100
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []SellHistoryItem
	for rows.Next() {
		var h SellHistoryItem
		var ts sql.NullString
		if err := rows.Scan(&h.ID, &h.ProductName, &h.Barcode, &h.Quantity, &h.Price, &ts); err != nil {
			return nil, err
		}
		h.Total = h.Price * float64(h.Quantity)
		if ts.Valid {
			h.Timestamp, h.TimestampRaw = formatTimestamp(ts.String)
		}
		result = append(result, h)
	}
	return result, rows.Err()
}

// CreateReturn — оформить возврат товара
func CreateReturn(username string, historyID, productID, qty int, productName, barcode, note string, price float64) error {
	if qty <= 0 {
		return errors.New("Количество должно быть больше нуля")
	}

	// проверяем что в истории продажи есть такое количество
	var soldQty int
	err := database.DB.QueryRow(
		`SELECT quantity FROM history WHERE id = ? AND action = 'sell'`, historyID,
	).Scan(&soldQty)
	if err != nil {
		return errors.New("Запись продажи не найдена")
	}

	// проверяем сколько уже возвращено по этой продаже
	var alreadyReturned int
	database.DB.QueryRow(
		`SELECT COALESCE(SUM(quantity), 0) FROM returns WHERE history_id = ?`, historyID,
	).Scan(&alreadyReturned)

	if qty > soldQty-alreadyReturned {
		return errors.New("Нельзя вернуть больше чем было продано")
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// восстанавливаем остаток
	_, err = tx.Exec(
		`UPDATE products SET quantity = quantity + ? WHERE id = ?`, qty, productID,
	)
	if err != nil {
		return err
	}

	// записываем возврат
	_, err = tx.Exec(
		`INSERT INTO returns(username, history_id, product_id, product_name, barcode, quantity, price, note)
		 VALUES(?, ?, ?, ?, ?, ?, ?, ?)`,
		username, historyID, productID, productName, barcode, qty, price, note,
	)
	if err != nil {
		return err
	}

	// пишем в историю
	_, err = tx.Exec(
		`INSERT INTO history(action, username, target, barcode, quantity, price)
		 VALUES(?, ?, ?, ?, ?, ?)`,
		"return", username, productName, barcode, qty, price,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetReturns — список всех возвратов
func GetReturns() ([]ReturnItem, error) {
	rows, err := database.DB.Query(`
		SELECT
			r.id, r.username,
			u.first_name, u.last_name,
			r.history_id, r.product_id, r.product_name,
			r.barcode, r.quantity, r.price, r.note, r.timestamp
		FROM returns r
		LEFT JOIN users u ON r.username = u.username
		ORDER BY r.id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []ReturnItem
	for rows.Next() {
		var item ReturnItem
		var first, last sql.NullString
		var ts sql.NullString
		if err := rows.Scan(
			&item.ID, &item.Username,
			&first, &last,
			&item.HistoryID, &item.ProductID, &item.ProductName,
			&item.Barcode, &item.Quantity, &item.Price, &item.Note, &ts,
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
		item.Total = item.Price * float64(item.Quantity)
		if ts.Valid {
			item.Timestamp, item.TimestampRaw = formatTimestamp(ts.String)
		}
		result = append(result, item)
	}
	return result, rows.Err()
}
