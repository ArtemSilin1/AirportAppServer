package user

import (
	"AirPort/internal/handlers"
	"fmt"
	"net/http"

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
	router.POST("/users/registration", h.Register)
	router.POST("/users/login", h.Login)
}

func (h *Handler) Register(c *gin.Context) {
	var newUser Users

	if err := c.ShouldBindJSON(&newUser); err != nil {
		fmt.Printf("Ошибка при чтении данных JSON: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"})
		return
	}

	if len(newUser.Password) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "пароль должен быть не менее 6 символов"})
		return
	}

	token, err := newUser.RegisterUser(h.db)
	if err != nil {
		fmt.Printf("Ошибка при попытке регистрации: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *Handler) Login(c *gin.Context) {
	var loginUser Users
	if err := c.ShouldBindJSON(&loginUser); err != nil {
		fmt.Printf("Ошибка при чтении данных JSON: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"})
		return
	}

	token, err := loginUser.LoginUser(h.db)
	if err != nil {
		fmt.Printf("Ошибка при попытке входа в аккаунт: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "ошибка при авторизации"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
