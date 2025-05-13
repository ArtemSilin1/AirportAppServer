package control

import (
	"AirPort/internal/handlers"
	"AirPort/internal/handlers/user"
	"AirPort/package/logs"
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
	router.GET("/control/getTokens", h.GetTokens)
	router.POST("/control/generateToken", h.GenerateToken)
	router.POST("/control/checkValidToken", h.CheckValidToken)
}

func (h *Handler) GetTokens(c *gin.Context) {
	var tokensToGet Token

	tokens, err := tokensToGet.GetAllTokens(h.db)
	if err != nil {
		if logErr := logs.NewLog("Токен", "control", err); logErr != nil {
			fmt.Printf("Ошибка логирования: %s", logErr)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "внутренняя ошибка сервера"})
		return
	}
	c.JSON(http.StatusOK, tokens)
}

func (h *Handler) GenerateToken(c *gin.Context) {
	type RequestData struct {
		User  user.Users `json:"user"`
		Token Token      `json:"token"`
	}

	var requestData RequestData

	if err := c.ShouldBindJSON(&requestData); err != nil {
		fmt.Printf("Ошибка при чтении данных JSON: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"})
		return
	}

	err := requestData.User.CheckAccPassword(h.db)
	if err != nil {
		if err.Error() == "неверные данные" {
			if logErr := logs.NewLog("Токен", "control", err); logErr != nil {
				fmt.Printf("Ошибка логирования: %s", logErr)
			}
			c.JSON(http.StatusUnauthorized, gin.H{"error": "в доступе отказано"})
			return
		}
		fmt.Printf("Ошибка при попытке проверить пароль: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"})
		return
	}

	if err := requestData.Token.GenerateToken(h.db); err != nil {
		if logErr := logs.NewLog("Токен", "control", err); logErr != nil {
			fmt.Printf("Ошибка логирования: %s", logErr)
		}
		fmt.Printf("Ошибка при попытке проверить пароль: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Создано"})
}

func (h *Handler) CheckValidToken(c *gin.Context) {
	type RequestData struct {
		User  user.Users `json:"user"`
		Token Token      `json:"token"`
	}

	var requestData RequestData

	if err := c.ShouldBindJSON(&requestData); err != nil {
		fmt.Printf("Ошибка при чтении данных JSON: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"})
		return
	}

	access, err := requestData.Token.CheckValidToken(h.db)
	if err != nil {
		if logErr := logs.NewLog("Токен", "control", err); logErr != nil {
			fmt.Printf("Ошибка логирования: %s", logErr)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "ошибка при проверке токена"})
			return
		}
	}

	if access {
		newToken, err := requestData.User.UpdateUserRole(h.db)
		if err != nil {
			if logErr := logs.NewLog("Токен", "control", err); err != nil {
				fmt.Printf("Ошибка логирования: %s", logErr)
			}
			c.JSON(http.StatusOK, gin.H{"token": newToken})
			return
		}
	}
	c.JSON(http.StatusUnauthorized, gin.H{"message": "в доступе отказано"})
}
