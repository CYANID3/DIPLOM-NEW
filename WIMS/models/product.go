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
		"INSERT INTO history(action, username, target, barcode, quantity) VALUES(?, ?, ?, ?, ?)",
		"add", username, name, barcode, quantity,
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

	err = tx.QueryRow("SELECT name, barcode FROM products WHERE id=?", id).Scan(&name, &barcode)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM products WHERE id=?", id)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		"INSERT INTO history(action, username, target, barcode, quantity) VALUES(?, ?, ?, ?, ?)",
		"delete", username, name, barcode, 0,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func SellProduct(id int, qty int, username string) error {
	if qty <= 0 {
		return errors.New("invalid quantity")
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var name, barcode string
	var stock int

	err = tx.QueryRow("SELECT name, barcode, quantity FROM products WHERE id=?", id).
		Scan(&name, &barcode, &stock)
	if err != nil {
		return err
	}

	if stock < qty {
		return errors.New("not enough stock")
	}

	_, err = tx.Exec("UPDATE products SET quantity = quantity - ? WHERE id=?", qty, id)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		"INSERT INTO history(action, username, target, barcode, quantity) VALUES(?, ?, ?, ?, ?)",
		"sell", username, name, barcode, qty,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}
