package user

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type Users struct {
	Id       int    `db:"id" json:"id"`
	Username string `db:"username" json:"username"`
	Name     string `db:"name" json:"name"`
	Password string `db:"password" json:"password"`
	Email    string `db:"email" json:"email"`
	UserRole bool   `db:"userRole" json:"userRole"`
}

var secretJwt = []byte("t*8z#7Pk9Q$JmR!fX&LoVz2^BnYGaPqH1SbEw3CfU@XdTl%Vi0NjD4KuMrOeWsCg")

func (u *Users) GenerateJWT() (string, error) {
	claims := jwt.MapClaims{
		"username": u.Username,
		"email":    u.Email,
		"role":     u.UserRole,
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(secretJwt)
	if err != nil {
		return "", fmt.Errorf("Ошибка при создании токена: %s", err)
	}

	return signedToken, nil
}

func (u *Users) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("Ошибка при попытке хэшировать пароль: %s", err)
	}

	return string(hashedPassword), nil
}

func (u *Users) RegisterUser(db *pgxpool.Pool) (string, error) {
	ctx := context.Background()

	checkUserExist := `
		SELECT COUNT(*) 
		FROM Users 
		WHERE username = $1
	`

	var userCount int
	if err := db.QueryRow(ctx, checkUserExist, u.Username).Scan(&userCount); err != nil {
		return "", fmt.Errorf("Ошибка при проверке пользователя: %s", err)
	}
	if userCount != 0 {
		return "", fmt.Errorf("already exist")
	}

	InsertQuery := `
		INSERT INTO Users (username, name, password, email, userRole)
		VALUES
			($1, $2, $3, $4, $5)
		RETURNING id
	`

	var newUserId int
	hashedPassword, err := u.HashPassword(u.Password)
	if err != nil {
		return "", err
	}

	if err := db.QueryRow(
		ctx,
		InsertQuery,
		u.Username,
		u.Name,
		hashedPassword,
		u.Email,
		u.UserRole,
	).Scan(&newUserId); err != nil {
		return "", fmt.Errorf("Ошибка записи в бд при попытке регистрации: %s", err)
	}

	token, err := u.GenerateJWT()
	if err != nil {
		return "", err
	}

	return token, nil
}

func (u *Users) LoginUser(db *pgxpool.Pool) (string, error) {
	ctx := context.Background()

	query := `
   	SELECT id, username, name, password, email, userrole 
   	FROM Users 
   	WHERE username = $1
	`

	var dbUser Users
	err := db.QueryRow(ctx, query, u.Username).Scan(
		&dbUser.Id,
		&dbUser.Username,
		&dbUser.Name,
		&dbUser.Password,
		&dbUser.Email,
		&dbUser.UserRole,
	)

	if err != nil {
		return "", fmt.Errorf("Ошибка авторизации: %s", err)
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(dbUser.Password),
		[]byte(u.Password),
	); err != nil {
		return "", fmt.Errorf("Неверные данные")
	}

	u.Id = dbUser.Id
	u.Email = dbUser.Email
	u.UserRole = dbUser.UserRole

	token, err := dbUser.GenerateJWT()
	if err != nil {
		return "", err
	}

	return token, nil
}
