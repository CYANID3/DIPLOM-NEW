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
	Price       float64 // цена за единицу
	SoldTotal   float64 // сумма исходной продажи
	ReturnTotal float64 // сумма возврата (qty * price)
	RemainTotal float64 // остаток после возврата (sold - все возвраты)
	Note        string
	Timestamp   string
	TimestampRaw string
}

// SellHistoryItem — запись продажи для выбора при возврате
type SellHistoryItem struct {
	ID              int
	ProductName     string
	Barcode         string
	Quantity        int     // продано
	Price           float64 // цена за единицу
	Total           float64 // сумма продажи
	AlreadyReturned int     // уже возвращено
	CanReturn       int     // можно вернуть
	Timestamp       string
	TimestampRaw    string
}

// GetSellHistory — последние 100 продаж для выбора возврата.
// Количество уже возвращённых единиц вычисляется одним запросом через LEFT JOIN.
func GetSellHistory() ([]SellHistoryItem, error) {
	rows, err := database.DB.Query(`
		SELECT
			h.id, h.target, h.barcode, h.quantity, h.price, h.timestamp,
			COALESCE(SUM(r.quantity), 0) AS already_returned
		FROM history h
		LEFT JOIN returns r ON r.history_id = h.id
		WHERE h.action = 'sell'
		GROUP BY h.id
		ORDER BY h.id DESC
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
		if err := rows.Scan(
			&h.ID, &h.ProductName, &h.Barcode,
			&h.Quantity, &h.Price, &ts,
			&h.AlreadyReturned,
		); err != nil {
			return nil, err
		}
		h.Total = h.Price * float64(h.Quantity)
		h.CanReturn = h.Quantity - h.AlreadyReturned
		if ts.Valid {
			h.Timestamp, h.TimestampRaw = formatTimestamp(ts.String)
		}
		result = append(result, h)
	}
	return result, rows.Err()
}

// CreateReturn — оформить возврат товара.
func CreateReturn(username string, historyID, productID, qty int, productName, barcode, note string, price float64) error {
	if qty <= 0 {
		return errors.New("Количество должно быть больше нуля")
	}

	// проверяем что продажа существует и берём количество
	var soldQty int
	err := database.DB.QueryRow(
		`SELECT quantity FROM history WHERE id = ? AND action = 'sell'`, historyID,
	).Scan(&soldQty)
	if err != nil {
		return errors.New("Запись продажи не найдена")
	}

	// считаем сколько уже возвращено по этой продаже
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
	if _, err = tx.Exec(
		`UPDATE products SET quantity = quantity + ? WHERE id = ?`, qty, productID,
	); err != nil {
		return err
	}

	// записываем возврат
	if _, err = tx.Exec(
		`INSERT INTO returns(username, history_id, product_id, product_name, barcode, quantity, price, note)
		 VALUES(?, ?, ?, ?, ?, ?, ?, ?)`,
		username, historyID, productID, productName, barcode, qty, price, note,
	); err != nil {
		return err
	}

	// пишем в историю
	if _, err = tx.Exec(
		`INSERT INTO history(action, username, target, barcode, quantity, price)
		 VALUES(?, ?, ?, ?, ?, ?)`,
		"return", username, productName, barcode, qty, price,
	); err != nil {
		return err
	}

	return tx.Commit()
}

// GetReturns — список всех возвратов.
// Финансовые показатели вычисляются одним JOIN-запросом без N+1.
func GetReturns() ([]ReturnItem, error) {
	rows, err := database.DB.Query(`
		SELECT
			r.id, r.username,
			u.first_name, u.last_name,
			r.history_id, r.product_id, r.product_name,
			r.barcode, r.quantity, r.price, r.note, r.timestamp,
			COALESCE(h.quantity, 0)          AS sold_qty,
			COALESCE(agg.total_returned, 0)  AS total_returned
		FROM returns r
		LEFT JOIN users u ON r.username = u.username
		LEFT JOIN history h ON h.id = r.history_id
		LEFT JOIN (
			SELECT history_id, SUM(quantity) AS total_returned
			FROM returns
			GROUP BY history_id
		) agg ON agg.history_id = r.history_id
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
		var soldQty, totalReturned int

		if err := rows.Scan(
			&item.ID, &item.Username,
			&first, &last,
			&item.HistoryID, &item.ProductID, &item.ProductName,
			&item.Barcode, &item.Quantity, &item.Price, &item.Note, &ts,
			&soldQty, &totalReturned,
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

		item.ReturnTotal = item.Price * float64(item.Quantity)
		item.SoldTotal   = item.Price * float64(soldQty)
		item.RemainTotal = item.Price * float64(soldQty-totalReturned)

		if ts.Valid {
			item.Timestamp, item.TimestampRaw = formatTimestamp(ts.String)
		}
		result = append(result, item)
	}
	return result, rows.Err()
}
