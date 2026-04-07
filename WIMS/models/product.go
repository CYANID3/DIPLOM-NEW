package models

import (
	"database/sql"
	"errors"
	"fmt"
	"wims/database"
)

type Product struct {
	ID       int
	Name     string
	Barcode  string
	Price    float64
	Quantity int
	Total    float64
	Category string
	MinStock int
}

func itoa(n int) string     { return fmt.Sprintf("%d", n) }
func ftoa(f float64) string { return fmt.Sprintf("%.2f", f) }

func GetAllProducts() ([]Product, error) {
	rows, err := database.DB.Query(
		`SELECT id, name, barcode, price, quantity, category, min_stock FROM products`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Barcode, &p.Price,
			&p.Quantity, &p.Category, &p.MinStock,
		); err != nil {
			return nil, err
		}
		p.Total = p.Price * float64(p.Quantity)
		products = append(products, p)
	}
	return products, rows.Err()
}

func GetProductByID(id int) (*Product, error) {
	var p Product
	err := database.DB.QueryRow(
		`SELECT id, name, barcode, price, quantity, category, min_stock
		 FROM products WHERE id = ?`, id,
	).Scan(&p.ID, &p.Name, &p.Barcode, &p.Price, &p.Quantity, &p.Category, &p.MinStock)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	p.Total = p.Price * float64(p.Quantity)
	return &p, nil
}

func GetCategories() ([]string, error) {
	rows, err := database.DB.Query(
		`SELECT DISTINCT category FROM products WHERE category != '' ORDER BY category`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cats []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, rows.Err()
}

func CreateProduct(name, barcode, category string, price float64, quantity, minStock int, username string) error {
	if quantity <= 0 {
		return errors.New("Количество должно быть больше нуля")
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var id int
	err = tx.QueryRow(`SELECT id FROM products WHERE name = ?`, name).Scan(&id)

	if err == nil {
		_, err = tx.Exec(
			`UPDATE products SET quantity=quantity+?, price=?, barcode=?,
			 category=?, min_stock=? WHERE name=?`,
			quantity, price, barcode, category, minStock, name,
		)
	} else if err == sql.ErrNoRows {
		_, err = tx.Exec(
			`INSERT INTO products(name, barcode, price, quantity, category, min_stock)
			 VALUES(?, ?, ?, ?, ?, ?)`,
			name, barcode, price, quantity, category, minStock,
		)
	} else {
		return err
	}
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		`INSERT INTO history(action, username, target, barcode, quantity, price)
		 VALUES(?, ?, ?, ?, ?, ?)`,
		"add", username, name, barcode, quantity, price,
	)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func UpdateProduct(id int, name, barcode, category string, price float64, minStock int) error {
	_, err := database.DB.Exec(
		`UPDATE products SET name=?, barcode=?, price=?, category=?, min_stock=? WHERE id=?`,
		name, barcode, price, category, minStock, id,
	)
	return err
}

func DeleteProduct(id int, username string) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var name, barcode string
	var quantity int
	var price float64
	err = tx.QueryRow(
		`SELECT name, barcode, quantity, price FROM products WHERE id=?`, id,
	).Scan(&name, &barcode, &quantity, &price)
	if err != nil {
		return err
	}

	if _, err = tx.Exec(`DELETE FROM products WHERE id=?`, id); err != nil {
		return err
	}

	_, err = tx.Exec(
		`INSERT INTO history(action, username, target, barcode, quantity, price)
		 VALUES(?, ?, ?, ?, ?, ?)`,
		"delete", username, name, barcode, quantity, price,
	)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func SellProduct(id, qty int, username string) error {
	if qty <= 0 {
		return errors.New("Количество должно быть больше нуля")
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var name, barcode string
	var stock int
	var price float64
	err = tx.QueryRow(
		`SELECT name, barcode, quantity, price FROM products WHERE id=?`, id,
	).Scan(&name, &barcode, &stock, &price)
	if err != nil {
		return err
	}

	if stock < qty {
		return errors.New("Недостаточно товара на складе")
	}

	if _, err = tx.Exec(
		`UPDATE products SET quantity=quantity-? WHERE id=?`, qty, id,
	); err != nil {
		return err
	}

	_, err = tx.Exec(
		`INSERT INTO history(action, username, target, barcode, quantity, price)
		 VALUES(?, ?, ?, ?, ?, ?)`,
		"sell", username, name, barcode, qty, price,
	)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func ExportProductsCSV() ([][]string, error) {
	rows, err := database.DB.Query(
		`SELECT id, name, barcode, category, price, quantity, min_stock
		 FROM products ORDER BY name`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := [][]string{
		{"ID", "Название", "Штрихкод", "Категория", "Цена", "Остаток", "Мин. остаток"},
	}
	for rows.Next() {
		var id, qty, minStock int
		var name, barcode, category string
		var price float64
		if err := rows.Scan(&id, &name, &barcode, &category, &price, &qty, &minStock); err != nil {
			return nil, err
		}
		result = append(result, []string{
			itoa(id), name, barcode, category, ftoa(price), itoa(qty), itoa(minStock),
		})
	}
	return result, rows.Err()
}

func ExportHistoryCSV() ([][]string, error) {
	rows, err := database.DB.Query(
		`SELECT id, action, username, target, barcode, quantity, price, timestamp
		 FROM history ORDER BY id DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := [][]string{
		{"ID", "Действие", "Пользователь", "Товар", "Штрихкод", "Кол-во", "Цена", "Время"},
	}
	for rows.Next() {
		var id, qty int
		var action, username, target, barcode, ts string
		var price float64
		if err := rows.Scan(&id, &action, &username, &target, &barcode, &qty, &price, &ts); err != nil {
			return nil, err
		}
		display, _ := formatTimestamp(ts)
		result = append(result, []string{
			itoa(id), action, username, target, barcode, itoa(qty), ftoa(price), display,
		})
	}
	return result, rows.Err()
}

// RestockProduct — пополняет остаток существующего товара
func RestockProduct(id, qty int, username string) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var name, barcode string
	var price float64
	err = tx.QueryRow(
		`SELECT name, barcode, price FROM products WHERE id=?`, id,
	).Scan(&name, &barcode, &price)
	if err != nil {
		return err
	}

	if _, err = tx.Exec(
		`UPDATE products SET quantity = quantity + ? WHERE id=?`, qty, id,
	); err != nil {
		return err
	}

	_, err = tx.Exec(
		`INSERT INTO history(action, username, target, barcode, quantity, price)
		 VALUES(?, ?, ?, ?, ?, ?)`,
		"restock", username, name, barcode, qty, price,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}
