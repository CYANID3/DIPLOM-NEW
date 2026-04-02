package models

import (
	"wims/database"
)

type Product struct {
	ID       int
	Name     string
	Price    float64
	Quantity int
	Total    float64 // итоговая стоимость = Price * Quantity
}

// Получение всех продуктов с расчетом Total
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

// Создание товара с записью в историю
func CreateProduct(name string, price float64, quantity int, username string) error {
	_, err := database.DB.Exec(
		"INSERT INTO products(name, price, quantity) VALUES(?, ?, ?)",
		name, price, quantity,
	)
	if err != nil {
		return err
	}

	_, err = database.DB.Exec(
		"INSERT INTO history(action, username, target, quantity) VALUES(?, ?, ?, ?)",
		"add", username, name, quantity,
	)
	return err
}

// Удаление товара с записью в историю
func DeleteProduct(id int, username string) error {
	var name string
	if err := database.DB.QueryRow("SELECT name FROM products WHERE id=?", id).Scan(&name); err != nil {
		return err
	}

	_, err := database.DB.Exec("DELETE FROM products WHERE id=?", id)
	if err != nil {
		return err
	}

	_, err = database.DB.Exec(
		"INSERT INTO history(action, username, target, quantity) VALUES(?, ?, ?, ?)",
		"delete", username, name, 0,
	)
	return err
}

// Продажа товара с записью в историю
func SellProduct(id int, qty int, username string) error {
	var name string
	var currentQty int
	if err := database.DB.QueryRow("SELECT name, quantity FROM products WHERE id=?", id).Scan(&name, &currentQty); err != nil {
		return err
	}

	if currentQty < qty {
		return nil // недостаточно товара
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

	_, err = database.DB.Exec(
		"INSERT INTO history(action, username, target, quantity) VALUES(?, ?, ?, ?)",
		"sell", username, name, qty,
	)
	return err
}
