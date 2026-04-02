package models

import (
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

// Создание пользователя (для админа)
func CreateUser(username, password, role, firstName, lastName, middleName, position, email string) error {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	_, err := database.DB.Exec(
		`INSERT INTO users(username, password, role, first_name, last_name, middle_name, position, email)
		 VALUES(?, ?, ?, ?, ?, ?, ?, ?)`,
		username, string(hash), role, firstName, lastName, middleName, position, email,
	)
	return err
}

// Проверка логина и пароля
func CheckUser(username, password string) (bool, *User) {
	var u User
	err := database.DB.QueryRow(
		"SELECT id, password, role, first_name, last_name, middle_name, position, email FROM users WHERE username = ?",
		username,
	).Scan(&u.ID, &u.Password, &u.Role, &u.FirstName, &u.LastName, &u.MiddleName, &u.Position, &u.Email)

	if err != nil {
		return false, nil
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil, &u
}

// Получение пользователя по username
func GetUserByUsername(username string) *User {
	var u User
	err := database.DB.QueryRow(
		"SELECT id, username, role, first_name, last_name, middle_name, position, email FROM users WHERE username = ?",
		username,
	).Scan(&u.ID, &u.Username, &u.Role, &u.FirstName, &u.LastName, &u.MiddleName, &u.Position, &u.Email)
	if err != nil {
		return &User{}
	}
	return &u
}

// Обновление профиля (для себя)
func UpdateProfile(username, firstName, lastName, middleName, position, email string) error {
	_, err := database.DB.Exec(
		"UPDATE users SET first_name=?, last_name=?, middle_name=?, position=?, email=? WHERE username=?",
		firstName, lastName, middleName, position, email, username,
	)
	return err
}

// Обновление пользователя (для админа)
func UpdateUser(username, firstName, lastName, middleName, position, email, role string) error {
	_, err := database.DB.Exec(
		"UPDATE users SET first_name=?, last_name=?, middle_name=?, position=?, email=?, role=? WHERE username=?",
		firstName, lastName, middleName, position, email, role, username,
	)
	return err
}

// Получение всех пользователей
func GetAllUsers() ([]User, error) {
	rows, err := database.DB.Query(
		"SELECT id, username, role, first_name, last_name, middle_name, position, email FROM users ORDER BY id",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		rows.Scan(&u.ID, &u.Username, &u.Role, &u.FirstName, &u.LastName, &u.MiddleName, &u.Position, &u.Email)
		users = append(users, u)
	}
	return users, nil
}

// Удаление пользователя
func DeleteUser(username string) error {
	_, err := database.DB.Exec("DELETE FROM users WHERE username=?", username)
	return err
}
