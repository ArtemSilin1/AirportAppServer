package user

import (
	"context"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type SecretJwt struct {
	Secret string `env:SECRET_JWT`
}

func (s *SecretJwt) ReadSecret() error {
	err := cleanenv.ReadConfig("internal/config/.env", s)
	if err != nil {
		fmt.Printf("Ошибка при получении секрета: %w", err)
		return err
	}

	return nil
}

func (u *Users) GenerateJWT() (string, error) {
	var Secret SecretJwt
	if err := Secret.ReadSecret(); err != nil {
		fmt.Println("Ошибка при попытке прочитать секрет: %w", err)
		return "", err
	}

	claims := jwt.MapClaims{
		"username": u.Username,
		"email":    u.Email,
		"role":     u.UserRole,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(Secret.Secret))
	if err != nil {
		return "", fmt.Errorf("Ошибка при создании токена: %w", err)
	}

	return signedToken, nil
}

func (u *Users) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("Ошибка при попытке хэшировать пароль: %w", err)
	}

	return string(hashedPassword), nil
}

type Users struct {
	Id       int    `db:"id" json:"id"`
	Username string `db:"username" json:"username"`
	Name     string `db:"name" json:"name"`
	Password string `db:"password" json:"password"`
	Email    string `db:"email" json:"email"`
	UserRole bool   `db:"userRole" json:"userRole"`
}

func (u *Users) RegisterUser(db *pgxpool.Pool) (string, error) {
	ctx := context.Background()

	checkUsernameExist := `
		SELECT COUNT(*) 
		FROM Users 
		WHERE username = $1
	`

	var userCountWithUsername int
	if err := db.QueryRow(ctx, checkUsernameExist, u.Username).Scan(&userCountWithUsername); err != nil {
		return "", fmt.Errorf("Ошибка при проверке пользователя: %w", err)
	}
	if userCountWithUsername != 0 {
		return "", fmt.Errorf(usernameAlreadyExistError)
	}

	checkEmailExist := `
		SELECT COUNT(*) 
		FROM Users 
		WHERE email = $1
	`

	var userCountWithEmail int
	if err := db.QueryRow(ctx, checkEmailExist, u.Email).Scan(&userCountWithEmail); err != nil {
		return "", fmt.Errorf("Ошибка при проверке пользователя: %w", err)
	}
	if userCountWithEmail != 0 {
		return "", fmt.Errorf(emailAlreadyExistError)
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
		return "", fmt.Errorf("Ошибка записи в бд при попытке регистрации: %w", err)
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
		return "", fmt.Errorf("Ошибка авторизации: %w", err)
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

func (u *Users) DeleteUser(db *pgxpool.Pool) error {
	ctx := context.Background()

	query := `
		DELETE FROM Users
		WHERE id = $1
	`

	_, err := db.Exec(ctx, query, u.Id)
	if err != nil {
		return fmt.Errorf("Ошибка при удалении: %w", err)
	}

	return nil
}
