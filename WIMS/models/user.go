package models

import (
	"wims/database"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int
	Username  string
	Password  string
	Role      string
	FirstName string
	LastName  string
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
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}

	_, err = database.DB.Exec(
		`INSERT INTO users(username, password, role, first_name, last_name)
		 VALUES(?,?,?,?,?)`,
		username, hash, role, firstName, lastName,
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

	database.DB.QueryRow(
		"SELECT id, username, role, first_name, last_name FROM users WHERE username=?",
		username,
	).Scan(&u.ID, &u.Username, &u.Role, &u.FirstName, &u.LastName)

	return &u
}

func GetAllUsers() ([]User, error) {
	rows, _ := database.DB.Query("SELECT username, role, first_name, last_name FROM users")

	var users []User
	for rows.Next() {
		var u User
		rows.Scan(&u.Username, &u.Role, &u.FirstName, &u.LastName)
		users = append(users, u)
	}
	return users, nil
}

func DeleteUser(username string) {
	database.DB.Exec("DELETE FROM users WHERE username=?", username)
}
