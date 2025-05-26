package board

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Board struct {
	Id           int    `db:"id" json:"id"`
	FlightNumber string `db:"flightNumber" json:"flightNumber"`
	Appointment  string `db:"appointment" json:"appointment"`
	Departure    string `db:"departure" json:"departure"`
	Status       string `db:"status" json:"status"`
}

const (
	DefaultStatus  = "Регистрация"
	StatusCanceled = "Отменён"
	TimeFormat     = "2006-01-02 15:04:05"
)

func (b *Board) CreateBoardItem(db *pgxpool.Pool) error {
	if len(b.FlightNumber) != 6 {
		return fmt.Errorf("недопустимый номер рейса")
	}

	departureTime, err := time.Parse(TimeFormat, b.Departure)
	if err != nil {
		return fmt.Errorf("неверный формат времени отправления: %w", err)
	}

	ctx := context.Background()

	checkQuery := `
		SELECT COUNT(*) 
		FROM Board 
		WHERE flightNumber = $1
	`

	var flightNumberCount int
	if err := db.QueryRow(ctx, checkQuery, b.FlightNumber).Scan(&flightNumberCount); err != nil {
		return fmt.Errorf("ошибка при проверке рейса: %w", err)
	}
	if flightNumberCount != 0 {
		return fmt.Errorf("такой рейс уже существует")
	}

	query := `
		INSERT INTO Board (flightNumber, appointment, departure, status, status_change_time)
		VALUES ($1, $2, $3, 'Регистрация', NOW())
	`

	if _, err := db.Exec(
		ctx,
		query,
		b.FlightNumber,
		b.Appointment,
		departureTime,
	); err != nil {
		return fmt.Errorf("ошибка при добавлении рейса: %w", err)
	}

	return nil
}

func (b *Board) DeleteBoardItem(db *pgxpool.Pool) error {
	ctx := context.Background()

	query := `
		DELETE FROM Board
		WHERE id = $1
	`
	_, err := db.Exec(ctx, query, b.Id)
	if err != nil {
		return fmt.Errorf("ошибка при попытке удалить из бд: %s", err)
	}

	return nil
}

func (b *Board) GetBoard(db *pgxpool.Pool) ([]Board, error) {
	ctx := context.Background()

	query := `
		SELECT 
			id, 
			flightnumber, 
			appointment, 
			TO_CHAR(departure, 'YYYY-MM-DD HH24:MI:SS'), 
			status 
		FROM Board
	`

	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var boardRows []Board
	for rows.Next() {
		var boardItem Board
		if err := rows.Scan(
			&boardItem.Id,
			&boardItem.FlightNumber,
			&boardItem.Appointment,
			&boardItem.Departure,
			&boardItem.Status,
		); err != nil {
			return nil, err
		}
		boardRows = append(boardRows, boardItem)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return boardRows, nil
}

func (b *Board) ChangeFlightStatus(db *pgxpool.Pool) error {
	ctx := context.Background()

	query := `
		UPDATE Board
		SET
			status = $1,
			status_change_time = NOW()
		WHERE id = $2
	`

	_, err := db.Exec(ctx, query, b.Status, b.Id)
	if err != nil {
		return fmt.Errorf("ошибка при обновлении статуса: %w", err)
	}

	return nil
}

func (b *Board) SelectDeparturePoint(db *pgxpool.Pool) ([]string, error) {
	ctx := context.Background()

	query := `
		SELECT appointment 
		FROM Board
	`

	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var startLocations []string
	for rows.Next() {
		var boardItem Board
		if err := rows.Scan(&boardItem.Departure); err != nil {
			return nil, err
		}
		getLocationStart := strings.Split(boardItem.Departure, " ")
		if len(getLocationStart) > 0 {
			startLocations = append(startLocations, getLocationStart[0])
		}
	}

	return startLocations, nil
}

func (b *Board) SelectDepartureEndPoint(db *pgxpool.Pool, startLocation string) ([]string, error) {
	ctx := context.Background()

	query := `
		SELECT appointment 
		FROM Board
	`

	rows, err := db.Query(ctx, query, startLocation+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var endLocations []string
	for rows.Next() {
		var boardItem Board
		if err := rows.Scan(&boardItem.Departure); err != nil {
			return nil, err
		}
		getLocationEnd := strings.Split(boardItem.Departure, " ")
		if len(getLocationEnd) > 0 {
			endLocations = append(endLocations, getLocationEnd[0])
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return endLocations, nil
}

func (b *Board) SelectAllFlight(db *pgxpool.Pool) ([]Board, error) {
	ctx := context.Background()

	query := `
      SELECT id, appointment 
      FROM Board
      WHERE status = 'Регистрация'
    `

	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var startLocations []Board
	for rows.Next() {
		var boardItem Board
		if err := rows.Scan(&boardItem.Id, &boardItem.Appointment); err != nil {
			return nil, err
		}
		startLocations = append(startLocations, boardItem)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return startLocations, nil
}
