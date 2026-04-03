package models

import (
	"database/sql"
	"errors"
	"wims/database"
)

type Product struct {
	ID       int
	Name     string
	Price    float64
	Quantity int
	Total    float64
}

// Получение всех продуктов
func GetAllProducts() ([]Product, error) {
	rows, err := database.DB.Query("SELECT id, name, price, quantity FROM products")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []Product

	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Quantity); err != nil {
			return nil, err
		}
		p.Total = p.Price * float64(p.Quantity)
		products = append(products, p)
	}

	return products, nil
}

// Создание или обновление товара
func CreateProduct(name string, price float64, quantity int, username string) error {
	if quantity <= 0 {
		return errors.New("quantity must be > 0")
	}

	var id int
	err := database.DB.QueryRow(
		"SELECT id FROM products WHERE name = ?",
		name,
	).Scan(&id)

	if err == nil {
		// обновление
		_, err = database.DB.Exec(
			"UPDATE products SET quantity = quantity + ?, price = ? WHERE name = ?",
			quantity, price, name,
		)
		if err != nil {
			return err
		}
	} else if err == sql.ErrNoRows {
		// создание
		_, err = database.DB.Exec(
			"INSERT INTO products(name, price, quantity) VALUES(?, ?, ?)",
			name, price, quantity,
		)
		if err != nil {
			return err
		}
	} else {
		return err
	}

	// ВАЖНО: сохраняем username (логин), а не ФИО
	_, err = database.DB.Exec(
		"INSERT INTO history(action, username, target, quantity) VALUES(?, ?, ?, ?)",
		"add", username, name, quantity,
	)

	return err
}

// Удаление товара
func DeleteProduct(id int, username string) error {
	var name string

	err := database.DB.QueryRow("SELECT name FROM products WHERE id=?", id).Scan(&name)
	if err != nil {
		return err
	}

	_, err = database.DB.Exec("DELETE FROM products WHERE id=?", id)
	if err != nil {
		return err
	}

	// ВАЖНО: сохраняем username
	_, err = database.DB.Exec(
		"INSERT INTO history(action, username, target, quantity) VALUES(?, ?, ?, ?)",
		"delete", username, name, 0,
	)

	return err
}

// Продажа товара
func SellProduct(id int, qty int, username string) error {
	if qty <= 0 {
		return errors.New("quantity must be > 0")
	}

	var name string
	var currentQty int

	err := database.DB.QueryRow(
		"SELECT name, quantity FROM products WHERE id=?",
		id,
	).Scan(&name, &currentQty)

	if err != nil {
		return err
	}

	if currentQty < qty {
		return errors.New("not enough stock")
	}

	res, err := database.DB.Exec(
		"UPDATE products SET quantity = quantity - ? WHERE id = ?",
		qty, id,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("update failed")
	}

	// ВАЖНО: сохраняем username
	_, err = database.DB.Exec(
		"INSERT INTO history(action, username, target, quantity) VALUES(?, ?, ?, ?)",
		"sell", username, name, qty,
	)

	return err
}
