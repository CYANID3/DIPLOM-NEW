package models

import (
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
	var id int

	err := database.DB.QueryRow(
		"SELECT id FROM products WHERE name = ?",
		name,
	).Scan(&id)

	if err == nil {
		// существует -> обновляем
		_, err = database.DB.Exec(
			"UPDATE products SET quantity = quantity + ?, price = ? WHERE name = ?",
			quantity, price, name,
		)
		if err != nil {
			return err
		}
	} else {
		// нет -> создаём
		_, err = database.DB.Exec(
			"INSERT INTO products(name, price, quantity) VALUES(?, ?, ?)",
			name, price, quantity,
		)
		if err != nil {
			return err
		}
	}

	displayName := GetDisplayName(username)

	_, err = database.DB.Exec(
		"INSERT INTO history(action, username, target, quantity) VALUES(?, ?, ?, ?)",
		"add", displayName, name, quantity,
	)
	return err
}

// Удаление товара
func DeleteProduct(id int, username string) error {
	var name string

	if err := database.DB.QueryRow("SELECT name FROM products WHERE id=?", id).Scan(&name); err != nil {
		return err
	}

	_, err := database.DB.Exec("DELETE FROM products WHERE id=?", id)
	if err != nil {
		return err
	}

	displayName := GetDisplayName(username)

	_, err = database.DB.Exec(
		"INSERT INTO history(action, username, target, quantity) VALUES(?, ?, ?, ?)",
		"delete", displayName, name, 0,
	)
	return err
}

// Продажа товара
func SellProduct(id int, qty int, username string) error {
	var name string
	var currentQty int

	if err := database.DB.QueryRow("SELECT name, quantity FROM products WHERE id=?", id).Scan(&name, &currentQty); err != nil {
		return err
	}

	if currentQty < qty {
		return nil
	}

	res, err := database.DB.Exec(
		"UPDATE products SET quantity = quantity - ? WHERE id = ?",
		qty, id,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil || rowsAffected == 0 {
		return nil
	}

	displayName := GetDisplayName(username)

	_, err = database.DB.Exec(
		"INSERT INTO history(action, username, target, quantity) VALUES(?, ?, ?, ?)",
		"sell", displayName, name, qty,
	)
	return err
}
