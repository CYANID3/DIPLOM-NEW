package models

import (
	"database/sql"
	"errors"
	"wims/database"
)

type Product struct {
	ID       int
	Name     string
	Barcode  string
	Price    float64
	Quantity int
	Total    float64
}

func GetAllProducts() ([]Product, error) {
	rows, err := database.DB.Query("SELECT id, name, barcode, price, quantity FROM products")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []Product

	for rows.Next() {
		var p Product
		err := rows.Scan(&p.ID, &p.Name, &p.Barcode, &p.Price, &p.Quantity)
		if err != nil {
			return nil, err
		}
		p.Total = p.Price * float64(p.Quantity)
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

func CreateProduct(name, barcode string, price float64, quantity int, username string) error {
	if quantity <= 0 {
		return errors.New("invalid quantity")
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var id int
	err = tx.QueryRow("SELECT id FROM products WHERE name = ?", name).Scan(&id)

	if err == nil {
		_, err = tx.Exec(
			"UPDATE products SET quantity = quantity + ?, price = ?, barcode = ? WHERE name = ?",
			quantity, price, barcode, name,
		)
	} else if err == sql.ErrNoRows {
		_, err = tx.Exec(
			"INSERT INTO products(name, barcode, price, quantity) VALUES(?, ?, ?, ?)",
			name, barcode, price, quantity,
		)
	} else {
		return err
	}

	if err != nil {
		return err
	}

	_, err = tx.Exec(
		"INSERT INTO history(action, username, target, barcode, quantity, price) VALUES(?, ?, ?, ?, ?, ?)",
		"add", username, name, barcode, quantity, price,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
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
		"SELECT name, barcode, quantity, price FROM products WHERE id=?",
		id,
	).Scan(&name, &barcode, &quantity, &price)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM products WHERE id=?", id)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		"INSERT INTO history(action, username, target, barcode, quantity, price) VALUES(?, ?, ?, ?, ?, ?)",
		"delete", username, name, barcode, quantity, price,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func SellProduct(id int, qty int, username string) error {
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

	err = tx.QueryRow("SELECT name, barcode, quantity, price FROM products WHERE id=?", id).
		Scan(&name, &barcode, &stock, &price)
	if err != nil {
		return err
	}

	if stock < qty {
		return errors.New("Недостаточно товара на складе")
	}

	_, err = tx.Exec("UPDATE products SET quantity = quantity - ? WHERE id=?", qty, id)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		"INSERT INTO history(action, username, target, barcode, quantity, price) VALUES(?, ?, ?, ?, ?, ?)",
		"sell", username, name, barcode, qty, price,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}
