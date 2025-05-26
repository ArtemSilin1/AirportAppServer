package report

import (
	"AirPort/internal/handlers"
	"AirPort/package/logs"
	"fmt"
	"net/http"
	"os"

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
	router.POST("/report/generateReport", h.GenerateReport)
}

func (h *Handler) GenerateReport(c *gin.Context) {
	var report Report

	if err := c.ShouldBindJSON(&report); err != nil {
		fmt.Println("ошибка чтения данных JSON")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"})
		return
	}

	if err := report.GetNewReportData(h.db); err != nil {
		if logErr := logs.NewLog("Отчёт", "report", err); logErr != nil {
			fmt.Printf("ошибка логирования: %s", logErr)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"})
		return
	}

	filePath := "./reports/report.pdf"

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "отчёт не найден"})
		return
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename=report.pdf")
	c.Header("Content-Type", "application/pdf")

	c.File(filePath)
}
