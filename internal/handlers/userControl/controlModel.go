package control

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

type Token struct {
	Id        int    `db:"id" json:"id"`
	Token     string `db: "token" json:"masterToken"`
	AddedDate string `db: "addedDate" json:"addedDate"`
}

func (t *Token) GenerateToken(db *pgxpool.Pool) error {
	length := rand.Intn(5) + 4
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	tokenStr := string(b)

	// --
	ctx := context.Background()

	query := `
      INSERT INTO Master_Tokens(token)
      VALUES ($1)
   `

	_, err := db.Exec(ctx, query, tokenStr)
	if err != nil {
		return fmt.Errorf("ошибка при вставке токена: %w", err)
	}

	return nil
}

func (t *Token) GetAllTokens(db *pgxpool.Pool) ([]string, error) {
	ctx := context.Background()

	query := `
      SELECT token FROM Master_Tokens
   `

	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении токенов: %s", err)
	}
	defer rows.Close()

	var tokens []string
	for rows.Next() {
		var token string
		if err := rows.Scan(&token); err != nil {
			return nil, fmt.Errorf("ошибка сканирования токена: %s", err)
		}
		tokens = append(tokens, token)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при обработке результатов: %s", err)
	}

	return tokens, nil
}

func (t *Token) CheckValidToken(db *pgxpool.Pool) (bool, error) {
	ctx := context.Background()

	query := `
		SELECT COUNT(*) FROM Master_Tokens
		WHERE token = $1
	`

	var tokenCount int

	if err := db.QueryRow(ctx, query, t.Token).Scan(&tokenCount); err != nil {
		return false, fmt.Errorf("ошибка при проверке токена")
	}

	if tokenCount == 0 {
		return false, nil
	}

	deleteQuery := `
		DELETE FROM Master_Tokens
		WHERE token = $1
	`

	_, err := db.Exec(ctx, deleteQuery, t.Token)
	if err != nil {
		return false, fmt.Errorf("ошибка при удалении токена")
	}

	return true, nil
}
