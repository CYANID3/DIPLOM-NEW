package models

import (
	"errors"
	"wims/database"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID         int
	Username   string
	Password   string
	Role       string
	FirstName  string
	LastName   string
	MiddleName string
	Position   string
	Email      string
}

func UpdateProfile(username, firstName, lastName, middleName, position, email string) error {
	_, err := database.DB.Exec(
		`UPDATE users 
		 SET first_name=?, last_name=?, middle_name=?, position=?, email=? 
		 WHERE username=?`,
		firstName, lastName, middleName, position, email, username,
	)
	return err
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func CreateUser(username, password, role, firstName, lastName, middleName, position, email string) error {
	if username == "" || password == "" {
		return errors.New("empty username or password")
	}

	hash, err := HashPassword(password)
	if err != nil {
		return err
	}

	_, err = database.DB.Exec(
		`INSERT INTO users(username, password, role, first_name, last_name, middle_name, position, email)
		 VALUES(?,?,?,?,?,?,?,?)`,
		username, hash, role, firstName, lastName, middleName, position, email,
	)

	return err
}

func CheckUser(username, password string) (bool, *User) {
	var u User

	err := database.DB.QueryRow(
		"SELECT id, password, role FROM users WHERE username=?",
		username,
	).Scan(&u.ID, &u.Password, &u.Role)

	if err != nil {
		return false, nil
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil, &u
}

func CheckPassword(username, password string) (bool, error) {
	var hash string

	err := database.DB.QueryRow(
		"SELECT password FROM users WHERE username=?",
		username,
	).Scan(&hash)

	if err != nil {
		return false, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil, nil
}

func UpdatePassword(username, password string) error {
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}

	_, err = database.DB.Exec(
		"UPDATE users SET password=? WHERE username=?",
		hash, username,
	)

	return err
}

func GetUserByUsername(username string) *User {
	var u User

	err := database.DB.QueryRow(
		`SELECT id, username, role, first_name, last_name, middle_name, position, email 
		 FROM users WHERE username=?`,
		username,
	).Scan(
		&u.ID, &u.Username, &u.Role,
		&u.FirstName, &u.LastName,
		&u.MiddleName, &u.Position, &u.Email,
	)

	if err != nil {
		return nil
	}

	return &u
}

func GetAllUsers() ([]User, error) {
	rows, err := database.DB.Query(
		"SELECT username, role, first_name, last_name FROM users",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User

	for rows.Next() {
		var u User
		err := rows.Scan(&u.Username, &u.Role, &u.FirstName, &u.LastName)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func DeleteUser(username string) error {
	_, err := database.DB.Exec("DELETE FROM users WHERE username=?", username)
	return err
}
