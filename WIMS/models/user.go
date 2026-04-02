package models

import (
	"wims/database"

	"golang.org/x/crypto/bcrypt"
)

// User структура пользователя
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

// Хэширование пароля
func HashPassword(password string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash)
}

// Создание пользователя
func CreateUser(username, password, role, firstName, lastName, middleName, position, email string) error {
	hashed := HashPassword(password)

	// По умолчанию роль "user", если пустая
	if role == "" {
		role = "user"
	}

	_, err := database.DB.Exec(
		`INSERT INTO users(username, password, role, first_name, last_name, middle_name, position, email)
		 VALUES(?, ?, ?, ?, ?, ?, ?, ?)`,
		username, hashed, role, firstName, lastName, middleName, position, email,
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
		return &User{Role: "user"} // по умолчанию роль user
	}
	return &u
}

// Обновление профиля (сам пользователь)
func UpdateProfile(username, firstName, lastName, middleName, position, email string) error {
	_, err := database.DB.Exec(
		"UPDATE users SET first_name=?, last_name=?, middle_name=?, position=?, email=? WHERE username=?",
		firstName, lastName, middleName, position, email, username,
	)
	return err
}

// Обновление пользователя (админ)
func UpdateUser(username, firstName, lastName, middleName, position, email, role string) error {
	// по умолчанию роль user
	if role == "" {
		role = "user"
	}
	_, err := database.DB.Exec(
		"UPDATE users SET first_name=?, last_name=?, middle_name=?, position=?, email=?, role=? WHERE username=?",
		firstName, lastName, middleName, position, email, role, username,
	)
	return err
}

// Обновление пароля пользователя (админ)
func UpdateUserPassword(username, password string) error {
	hashed := HashPassword(password)
	_, err := database.DB.Exec("UPDATE users SET password=? WHERE username=?", hashed, username)
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

func GetDisplayName(username string) string {
	var firstName, lastName string

	err := database.DB.QueryRow(
		"SELECT first_name, last_name FROM users WHERE username = ?",
		username,
	).Scan(&firstName, &lastName)

	if err != nil {
		return username
	}

	if firstName != "" || lastName != "" {
		return firstName + " " + lastName
	}

	return username
}
