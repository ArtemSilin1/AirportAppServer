package tickets

import (
	"AirPort/internal/handlers"
	"AirPort/package/logs"
	"net/http"

	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Handler struct {
	db *pgxpool.Pool
}

func NewHandler(db *pgxpool.Pool) handlers.Handlers {
	return &Handler{db: db}
}

func (h *Handler) RegisterHandler(router *gin.Engine) {
	router.POST("/ticket/getUserTickets", h.GetUserTickets)
	router.POST("/ticket/createUserTickets", h.CreateUserTicket)
}

func (h *Handler) GetUserTickets(c *gin.Context) {
	var request struct {
		UserId int `json:"user_id"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		logs.NewLog("Ticket", "ticket", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	ticket := Ticket{UserId: request.UserId}
	userTickets, err := ticket.GetAllUserTickets(h.db)
	if err != nil {
		logs.NewLog("Ticket", "ticket", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if len(userTickets) == 0 {
		c.JSON(http.StatusOK, gin.H{"rows": []interface{}{}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"rows": userTickets})
}

func (h *Handler) CreateUserTicket(c *gin.Context) {
	var Ticket Ticket
	if err := c.ShouldBindJSON(&Ticket); err != nil {
		if logErr := logs.NewLog("Билет", "ticket", err); logErr != nil {
			fmt.Printf("Ошибка логирования: %s", logErr)
		}
		fmt.Printf("Ошибка при чтении данных JSON: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"})
		return
	}

	if err := Ticket.CreateNewTicket(h.db); err != nil {
		if logErr := logs.NewLog("Билет", "ticket", err); logErr != nil {
			fmt.Printf("Ошибка логирования: %s", logErr)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Создано"})
}
