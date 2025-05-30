package user

import (
	"context"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

var (
	TimeFormat = "2006-01-02 15:04:05"
)

type SecretJwt struct {
	Secret string `env:SECRET_JWT`
}

func (s *SecretJwt) ReadSecret() error {
	err := cleanenv.ReadConfig("internal/config/.env", s)
	if err != nil {
		return fmt.Errorf("ошибка при получении секрета: %w", err)
	}

	return nil
}

func (u *Users) GenerateJWT() (string, error) {
	var Secret SecretJwt
	if err := Secret.ReadSecret(); err != nil {
		fmt.Println("ошибка при попытке прочитать секрет: %w", err)
		return "", err
	}

	claims := jwt.MapClaims{
		"id":          u.Id,
		"username":    u.Username,
		"email":       u.Email,
		"role":        u.UserRole,
		"name":        u.Name,
		"masterAdmin": u.MasterAdmin,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(Secret.Secret))
	if err != nil {
		return "", fmt.Errorf("ошибка при создании токена: %w", err)
	}

	return signedToken, nil
}

func (u *Users) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("ошибка при попытке хэшировать пароль: %w", err)
	}

	return string(hashedPassword), nil
}

type Users struct {
	Id          int    `db:"id" json:"id"`
	Username    string `db:"username" json:"username"`
	Name        string `db:"name" json:"name"`
	Password    string `db:"password" json:"password"`
	Email       string `db:"email" json:"email"`
	UserRole    bool   `db:"userRole" json:"userRole"`
	MasterAdmin bool   `db:"masterAdmin"`
}

func (u *Users) CheckAccPassword(db *pgxpool.Pool) error {
	ctx := context.Background()

	getPasswordQuery := `
		SELECT password
		FROM Users
		WHERE username = $1;
	`

	var dbUser Users
	if err := db.QueryRow(ctx, getPasswordQuery, u.Username).Scan(&dbUser.Password); err != nil {
		return fmt.Errorf("ошибка при попытке получить пароль из базы данных: %s", err)
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(dbUser.Password),
		[]byte(u.Password),
	); err != nil {
		return fmt.Errorf("неверные данные")
	}

	return nil
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
		return "", fmt.Errorf("ошибка при проверке пользователя: %w", err)
	}
	if userCountWithUsername != 0 {
		return "", fmt.Errorf("ошибка: %s", usernameAlreadyExistError)
	}

	checkEmailExist := `
		SELECT COUNT(*) 
		FROM Users 
		WHERE email = $1
	`

	var userCountWithEmail int
	if err := db.QueryRow(ctx, checkEmailExist, u.Email).Scan(&userCountWithEmail); err != nil {
		return "", fmt.Errorf("ошибка при проверке пользователя: %w", err)
	}
	if userCountWithEmail != 0 {
		return "", fmt.Errorf("ошибка: %s", emailAlreadyExistError)
	}

	InsertQuery := `
		INSERT INTO Users (username, name, password, email, userRole)
		VALUES
			($1, $2, $3, $4, false)
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
	).Scan(&newUserId); err != nil {
		return "", fmt.Errorf("ошибка записи в бд при попытке регистрации: %w", err)
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
   	SELECT id, username, name, password, email, userrole, masterAdmin 
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
		&dbUser.MasterAdmin,
	)

	if err != nil {
		return "", fmt.Errorf("ошибка авторизации: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(dbUser.Password),
		[]byte(u.Password),
	); err != nil {
		return "", fmt.Errorf("неверные данные")
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
		return fmt.Errorf("ошибка при удалении: %w", err)
	}

	return nil
}

func (u *Users) UpdateUserRole(db *pgxpool.Pool) (string, error) {
	ctx := context.Background()

	updateQuery := `
		UPDATE Users	
		SET userrole = true 
		WHERE username = $1
	`

	_, err := db.Exec(ctx, updateQuery, u.Username)
	if err != nil {
		return "", fmt.Errorf("ошибка при обновлении роли: %w", err)
	}

	selectQuery := `
		SELECT id, username, name, password, email, userrole 
		FROM Users 
		WHERE username = $1
	`

	var userNewRole Users
	if err := db.QueryRow(ctx, selectQuery, u.Username).Scan(
		&userNewRole.Id,
		&userNewRole.Username,
		&userNewRole.Name,
		&userNewRole.Password,
		&userNewRole.Email,
		&userNewRole.UserRole,
	); err != nil {
		return "", fmt.Errorf("ошибка при получении обновлённого пользователя: %w", err)
	}

	token, err := userNewRole.GenerateJWT()
	if err != nil {
		return "", fmt.Errorf("ошибка при генерации токена: %w", err)
	}

	return token, nil
}

type Notification struct {
	SeatNumber string
	Date       string
}

func (u *Users) GetAllNotifications(db *pgxpool.Pool) ([]Notification, error) {
	ctx := context.Background()

	query := `
	    SELECT Tickets.seatNumber, TO_CHAR(NOW(), 'YYYY-MM-DD HH24:MI:SS') as date
	    FROM Notifications
		JOIN Tickets ON Tickets.id = ticket_id
	    WHERE user_id = $1
	`

	rows, err := db.Query(ctx, query, u.Id)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения уведомлений: %w", err)
	}
	defer rows.Close()

	var notifications []Notification
	for rows.Next() {
		var noti Notification
		if err := rows.Scan(&noti.SeatNumber, &noti.Date); err != nil {
			return nil, fmt.Errorf("ошибка получения уведомлений: %w", err)
		}
		notifications = append(notifications, noti)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка получения массива: %w", err)
	}

	return notifications, nil
}
