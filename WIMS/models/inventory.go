package models

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
	"wims/database"
)

type Inventory struct {
	ID          int
	Username    string
	UserDisplay string
	Status      string
	Note        string
	CreatedAt   string
	CompletedAt string
	ItemCount   int
	DiffCount   int
}

type InventoryItemRow struct {
	ID          int
	InventoryID int
	ProductID   int
	ProductName string
	Barcode     string
	ExpectedQty int
	ActualQty   int
	Diff        int
	Price       float64
	Total       float64  // Price * ActualQty
	DiffTotal   float64  // Price * |Diff| — для акта
	Reason      string
}

// CreateInventory — создаёт новый документ инвентаризации
// автоматически заполняет строки всеми товарами с текущими остатками
func CreateInventory(username, note string) (int, error) {
	tx, err := database.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.Exec(
		`INSERT INTO inventories(username, status, note) VALUES(?, 'draft', ?)`,
		username, note,
	)
	if err != nil {
		return 0, err
	}
	invID, _ := res.LastInsertId()

	// заполняем строки всеми товарами
	rows, err := tx.Query(`SELECT id, name, barcode, quantity, price FROM products ORDER BY name`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var pid, qty int
		var name, barcode string
		var price float64
		if err := rows.Scan(&pid, &name, &barcode, &qty, &price); err != nil {
			return 0, err
		}
		if _, err := tx.Exec(
			`INSERT INTO inventory_items(inventory_id, product_id, product_name, barcode, expected_qty, actual_qty, diff, price)
			 VALUES(?, ?, ?, ?, ?, ?, 0, ?)`,
			invID, pid, name, barcode, qty, qty, price,
		); err != nil {
			return 0, err
		}
	}

	return int(invID), tx.Commit()
}

// GetInventories — список всех инвентаризаций
func GetInventories() ([]Inventory, error) {
	rows, err := database.DB.Query(`
		SELECT
			i.id, i.username,
			u.first_name, u.last_name,
			i.status, i.note, i.created_at,
			COALESCE(i.completed_at, ''),
			COUNT(ii.id),
			SUM(CASE WHEN ii.diff != 0 THEN 1 ELSE 0 END)
		FROM inventories i
		LEFT JOIN users u ON i.username = u.username
		LEFT JOIN inventory_items ii ON ii.inventory_id = i.id
		GROUP BY i.id
		ORDER BY i.id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Inventory
	for rows.Next() {
		var inv Inventory
		var first, last sql.NullString
		var createdRaw, completedRaw string
		var diffCount sql.NullInt64
		if err := rows.Scan(
			&inv.ID, &inv.Username,
			&first, &last,
			&inv.Status, &inv.Note, &createdRaw, &completedRaw,
			&inv.ItemCount, &diffCount,
		); err != nil {
			return nil, err
		}
		firstName := strings.TrimSpace(first.String)
		lastName  := strings.TrimSpace(last.String)
		full      := strings.TrimSpace(firstName + " " + lastName)
		if full != "" {
			inv.UserDisplay = full
		} else {
			inv.UserDisplay = inv.Username
		}
		inv.CreatedAt, _ = formatTimestamp(createdRaw)
		if completedRaw != "" {
			inv.CompletedAt, _ = formatTimestamp(completedRaw)
		}
		inv.DiffCount = int(diffCount.Int64)
		result = append(result, inv)
	}
	return result, rows.Err()
}

// GetInventoryByID — документ инвентаризации со строками
func GetInventoryByID(id int) (*Inventory, []InventoryItemRow, error) {
	var inv Inventory
	var first, last sql.NullString
	var createdRaw, completedRaw string

	err := database.DB.QueryRow(`
		SELECT i.id, i.username, u.first_name, u.last_name,
		       i.status, i.note, i.created_at, COALESCE(i.completed_at, '')
		FROM inventories i
		LEFT JOIN users u ON i.username = u.username
		WHERE i.id = ?`, id,
	).Scan(
		&inv.ID, &inv.Username, &first, &last,
		&inv.Status, &inv.Note, &createdRaw, &completedRaw,
	)
	if err != nil {
		return nil, nil, err
	}

	firstName := strings.TrimSpace(first.String)
	lastName  := strings.TrimSpace(last.String)
	full      := strings.TrimSpace(firstName + " " + lastName)
	if full != "" {
		inv.UserDisplay = full
	} else {
		inv.UserDisplay = inv.Username
	}
	inv.CreatedAt, _ = formatTimestamp(createdRaw)
	if completedRaw != "" {
		inv.CompletedAt, _ = formatTimestamp(completedRaw)
	}

	rows, err := database.DB.Query(`
		SELECT id, inventory_id, product_id, product_name, barcode,
		       expected_qty, actual_qty, diff, price, reason
		FROM inventory_items
		WHERE inventory_id = ?
		ORDER BY product_name`, id,
	)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var items []InventoryItemRow
	for rows.Next() {
		var item InventoryItemRow
		if err := rows.Scan(
			&item.ID, &item.InventoryID, &item.ProductID, &item.ProductName,
			&item.Barcode, &item.ExpectedQty, &item.ActualQty, &item.Diff, &item.Price, &item.Reason,
		); err != nil {
			return nil, nil, err
		}
		item.Total = item.Price * float64(item.ActualQty)
		diffAbs := item.Diff
		if diffAbs < 0 { diffAbs = -diffAbs }
		item.DiffTotal = item.Price * float64(diffAbs)
		items = append(items, item)
	}
	return &inv, items, rows.Err()
}

// UpdateInventoryItem — обновить фактическое количество по строке
func UpdateInventoryItem(itemID, actualQty int, reason string) error {
	_, err := database.DB.Exec(
		`UPDATE inventory_items SET actual_qty = ?, diff = ? - expected_qty, reason = ?
		 WHERE id = ?`, actualQty, actualQty, reason, itemID,
	)
	return err
}

// CompleteInventory — завершить инвентаризацию: скорректировать остатки
func CompleteInventory(id int, username string) error {
	inv, items, err := GetInventoryByID(id)
	if err != nil {
		return err
	}
	if inv.Status != "draft" {
		return fmt.Errorf("инвентаризация уже завершена")
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, item := range items {
		if item.Diff == 0 {
			continue
		}
		// корректируем остаток
		if _, err := tx.Exec(
			`UPDATE products SET quantity = ? WHERE id = ?`,
			item.ActualQty, item.ProductID,
		); err != nil {
			return err
		}
		// пишем в историю
		action := "surplus"   // излишек
		if item.Diff < 0 {
			action = "shortage" // недостача
		}
		qty := item.Diff
		if qty < 0 {
			qty = -qty
		}
		if _, err := tx.Exec(
			`INSERT INTO history(action, username, target, barcode, quantity, price)
			 VALUES(?, ?, ?, ?, ?, ?)`,
			action, username, item.ProductName, item.Barcode, qty, item.Price,
		); err != nil {
			return err
		}
	}

	now := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	if _, err := tx.Exec(
		`UPDATE inventories SET status = 'completed', completed_at = ? WHERE id = ?`,
		now, id,
	); err != nil {
		return err
	}

	return tx.Commit()
}

// ExportInventoryCSV — экспорт строк инвентаризации в CSV
func ExportInventoryCSV(id int) ([][]string, error) {
	_, items, err := GetInventoryByID(id)
	if err != nil {
		return nil, err
	}

	result := [][]string{
		{"ID", "Товар", "Штрихкод", "Ожидаемо", "Факт", "Разница", "Цена", "Сумма факт", "Причина расхождения"},
	}
	for _, item := range items {
		diff := fmt.Sprintf("%+d", item.Diff)
		result = append(result, []string{
			itoa(item.ID),
			item.ProductName,
			item.Barcode,
			itoa(item.ExpectedQty),
			itoa(item.ActualQty),
			diff,
			ftoa(item.Price),
			ftoa(item.Total),
			item.Reason,
		})
	}
	return result, nil
}
