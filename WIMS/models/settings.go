package models

import "wims/database"

// GetSetting возвращает значение настройки по ключу
func GetSetting(key string) string {
	var value string
	database.DB.QueryRow(
		`SELECT value FROM settings WHERE key = ?`, key,
	).Scan(&value)
	return value
}

// GetAllSettings возвращает все настройки как map
func GetAllSettings() map[string]string {
	rows, err := database.DB.Query(`SELECT key, value FROM settings`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var k, v string
		if rows.Scan(&k, &v) == nil {
			result[k] = v
		}
	}
	return result
}

// SetSetting сохраняет настройку
func SetSetting(key, value string) error {
	_, err := database.DB.Exec(
		`INSERT INTO settings(key, value) VALUES(?, ?)
		 ON CONFLICT(key) DO UPDATE SET value = excluded.value`,
		key, value,
	)
	return err
}
