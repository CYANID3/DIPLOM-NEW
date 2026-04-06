package models

import (
	"crypto/rand"
	"encoding/hex"
	"time"
	"wims/database"
)

type Session struct {
	Token     string
	Username  string
	CreatedAt string
	LastSeen  string
	UserAgent string
	IP        string
}

func CreateSession(username, userAgent, ip string) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)

	_, err := database.DB.Exec(
		`INSERT INTO sessions(token, username, user_agent, ip) VALUES(?, ?, ?, ?)`,
		token, username, userAgent, ip,
	)
	if err != nil {
		return "", err
	}
	return token, nil
}

func GetSession(token string) *Session {
	var s Session
	err := database.DB.QueryRow(
		`SELECT token, username, created_at, last_seen, user_agent, ip
		 FROM sessions WHERE token = ?`, token,
	).Scan(&s.Token, &s.Username, &s.CreatedAt, &s.LastSeen, &s.UserAgent, &s.IP)
	if err != nil {
		return nil
	}

	database.DB.Exec(
		`UPDATE sessions SET last_seen = ? WHERE token = ?`,
		time.Now().UTC().Format("2006-01-02T15:04:05Z"), token,
	)
	return &s
}

func DeleteSession(token string) {
	database.DB.Exec(`DELETE FROM sessions WHERE token = ?`, token)
}

func DeleteUserSessions(username string) {
	database.DB.Exec(`DELETE FROM sessions WHERE username = ?`, username)
}

func GetAllSessions() ([]Session, error) {
	rows, err := database.DB.Query(
		`SELECT s.token, s.username, s.created_at, s.last_seen, s.user_agent, s.ip
		 FROM sessions s
		 ORDER BY s.last_seen DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Session
	for rows.Next() {
		var s Session
		var createdRaw, lastSeenRaw string

		if err := rows.Scan(
			&s.Token, &s.Username, &createdRaw,
			&lastSeenRaw, &s.UserAgent, &s.IP,
		); err != nil {
			return nil, err
		}

		// форматируем даты через общую функцию из history.go
		s.CreatedAt, _ = formatTimestamp(createdRaw)
		s.LastSeen, _  = formatTimestamp(lastSeenRaw)

		result = append(result, s)
	}
	return result, rows.Err()
}
