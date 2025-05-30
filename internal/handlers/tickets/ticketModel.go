package tickets

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Ticket struct {
	Id         int    `json:"id" db:"id"`
	UserId     int    `json:"user_id" db:"userId"`
	FlightId   int    `json:"flight_id" db:"flightId"`
	SeatNumber string `json:"seat_number" db:"seatNumber"`
	Price      int    `json:"price" db:"price"`
}

func (t *Ticket) CreateNewTicket(db *pgxpool.Pool) error {
	randomTikcetPrice := rand.Intn(35000-19000+1) + 19000

	letters := []rune{'A', 'B', 'C'}
	letter := letters[rand.Intn(len(letters))]
	number := rand.Intn(21)
	seatNumber := fmt.Sprintf("%c%02d", letter, number)

	ctx := context.Background()

	query := `
		INSERT INTO Tickets(userId, flightId, seatNumber, price)
		VALUES
			($1, $2, $3, $4)
		RETURNING id
	`

	var ticket_id int

	err := db.QueryRow(ctx, query, t.UserId, t.FlightId, seatNumber, randomTikcetPrice).Scan(
		&ticket_id,
	)
	if err != nil {
		return err
	}

	addNotificationQuery := `
		INSERT INTO Notifications (user_id, ticket_id)
		VALUES
			($1, $2)
	`

	_, err = db.Exec(ctx, addNotificationQuery, t.UserId, ticket_id)
	if err != nil {
		return err
	}

	return nil
}

type UserTicketResponse struct {
	FlightNumber string `json:"flightId"`
	SeatNumber   string `json:"seatNumber"`
	Price        int    `json:"price"`
}

func (t *Ticket) GetAllUserTickets(db *pgxpool.Pool) ([]UserTicketResponse, error) {
	ctx := context.Background()

	query := `
      SELECT Board.flightNumber, Tickets.seatNumber, Tickets.price
      FROM Tickets
      JOIN Board ON Board.id = Tickets.flightId
      WHERE Tickets.userId = $1
   `

	rows, err := db.Query(ctx, query, t.UserId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var allTickets []UserTicketResponse
	for rows.Next() {
		var ticket UserTicketResponse
		if err := rows.Scan(&ticket.FlightNumber, &ticket.SeatNumber, &ticket.Price); err != nil {
			return nil, err
		}
		allTickets = append(allTickets, ticket)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return allTickets, nil
}
