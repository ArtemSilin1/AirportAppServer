package board

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
	router.GET("/board/getBoard", h.GetBoard)
	router.POST("board/createBoardItem", h.CreateBoardItem)
	router.PUT("board/updateBoardStatus", h.UpdateBoardStatus)
	router.DELETE("board/deleteFlight", h.DeleteFlight)
	router.GET("/board/getAllStartLocations", h.GetStartRoutes)
	router.POST("/board/getAllFinalLocations", h.GetEndRoutes)
}

func (h *Handler) GetBoard(c *gin.Context) {
	var boardToGet Board

	board, err := boardToGet.GetBoard(h.db)
	if err != nil {
		if logErr := logs.NewLog("Доска", "board", err); logErr != nil {
			fmt.Printf("Ошибка логирования: %s", logErr)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "внутренняя ошибка сервера"})
		return
	}

	if logErr := logs.NewLog("Доска", "board", nil); logErr != nil {
		fmt.Printf("Ошибка логирования: %s", logErr)
	}
	c.JSON(http.StatusOK, board)
}

func (h *Handler) CreateBoardItem(c *gin.Context) {
	type RequestData struct {
		User  user.Users `json:"user"`
		Board Board      `json:"board"`
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
			if logErr := logs.NewLog("Доска", "board", err); logErr != nil {
				fmt.Printf("Ошибка логирования: %s", logErr)
			}
			c.JSON(http.StatusUnauthorized, gin.H{"error": "в доступе отказано"})
			return
		}
		fmt.Printf("Ошибка при попытке изменить статус рейса: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"})
		return
	}

	if err := requestData.Board.CreateBoardItem(h.db); err != nil {
		if logErr := logs.NewLog("Доска", "board", err); logErr != nil {
			fmt.Printf("Ошибка логирования: %s", logErr)
		}
		fmt.Printf("Ошибка при попытке создать рейс: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Создано"})
}

func (h *Handler) UpdateBoardStatus(c *gin.Context) {
	type RequestData struct {
		User  user.Users `json:"user"`
		Board Board      `json:"board"`
	}

	var requestData RequestData

	if err := c.ShouldBindJSON(&requestData); err != nil {
		if logErr := logs.NewLog("Доска", "board", err); logErr != nil {
			fmt.Printf("Ошибка логирования: %s", logErr)
		}
		fmt.Printf("Ошибка при чтении данных JSON: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"})
		return
	}

	err := requestData.User.CheckAccPassword(h.db)
	if err != nil {
		if err.Error() == "неверные данные" {
			if logErr := logs.NewLog("Доска", "board", err); logErr != nil {
				fmt.Printf("Ошибка логирования: %s", logErr)
			}
			c.JSON(http.StatusUnauthorized, gin.H{"error": "в доступе отказано"})
			return
		}
		fmt.Printf("Ошибка при попытке изменить статус рейса: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"})
		return
	}

	if err := requestData.Board.ChangeFlightStatus(h.db); err != nil {
		if logErr := logs.NewLog("Доска", "board", err); logErr != nil {
			fmt.Printf("Ошибка логирования: %s", logErr)
		}
		fmt.Printf("Ошибка при попытке создать рейс: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "успешно"})
}

func (h *Handler) DeleteFlight(c *gin.Context) {
	type RequestData struct {
		User  user.Users `json:"user"`
		Board Board      `json:"board"`
	}

	var requestData RequestData

	if err := c.ShouldBindJSON(&requestData); err != nil {
		if logErr := logs.NewLog("Доска", "board", err); logErr != nil {
			fmt.Printf("Ошибка логирования: %s", logErr)
		}
		fmt.Printf("Ошибка при чтении данных JSON: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"})
		return
	}

	err := requestData.User.CheckAccPassword(h.db)
	if err != nil {
		if err.Error() == "неверные данные" {
			if logErr := logs.NewLog("Доска", "board", err); logErr != nil {
				fmt.Printf("Ошибка логирования: %s", logErr)
			}
			c.JSON(http.StatusUnauthorized, gin.H{"error": "в доступе отказано"})
			return
		}
		fmt.Printf("Ошибка при попытке удалить рейс: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"})
		return
	}

	if err := requestData.Board.DeleteBoardItem(h.db); err != nil {
		if logErr := logs.NewLog("Доска", "board", err); logErr != nil {
			fmt.Printf("Ошибка логирования: %s", logErr)
		}
		fmt.Printf("Ошибка при попытке удалить рейс: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "успешно"})
}

func (h *Handler) GetStartRoutes(c *gin.Context) {
	var board Board
	routes, err := board.SelectAllFlight(h.db)
	if err != nil {
		if logsErr := logs.NewLog("Board", "board", err); logsErr != nil {
			fmt.Printf("Ошибка логирования: %s", logsErr)
		}
		fmt.Printf("Ошибка при попытке получить список точек отправления: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"routes": routes})
}

func (h *Handler) GetEndRoutes(c *gin.Context) {
	var inputData struct {
		StartLocation string `json:"startLocation"`
	}

	if err := c.ShouldBindJSON(&inputData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "данные не прошли валидацию"})
		return
	}

	var routesToGet Board
	rows, err := routesToGet.SelectDepartureEndPoint(h.db, inputData.StartLocation)
	if err != nil {
		if logsErr := logs.NewLog("Board", "board", err); logsErr != nil {
			fmt.Printf("Ошибка логирования: %s", logsErr)
		}
		fmt.Printf("Ошибка при попытке получить список точек назначения: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"endPoints": rows})
}
