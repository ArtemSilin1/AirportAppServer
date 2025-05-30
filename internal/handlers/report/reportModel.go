package report

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/jung-kurt/gofpdf"
)

type Report struct {
	Interval int `json:"interval"`
}

func (r *Report) GetNewReportData(db *pgxpool.Pool) error {
	ctx := context.Background()

	selectedIntervalOption := ""

	switch r.Interval {
	case 1:
		selectedIntervalOption = "week"
	case 2:
		selectedIntervalOption = "month"
	case 3:
		selectedIntervalOption = "year"
	default:
		return fmt.Errorf("неподдерживаемый интвервал")
	}

	query := `
		SELECT 
			DATE_TRUNC($1, b.departure) AS period,
			COUNT(t.id) AS tickets_sold,
			SUM(t.price)::FLOAT AS daily_revenue,
			AVG(t.price)::FLOAT AS avg_price,
			AVG(COUNT(t.id)) OVER (
				ORDER BY DATE_TRUNC($1, b.departure) 
				ROWS BETWEEN 7 PRECEDING AND CURRENT ROW
			)::FLOAT AS moving_avg
		FROM 
			Board b
		JOIN 
			Tickets t ON b.id = t.flightId
		GROUP BY 
			period
		ORDER BY 
			period;
	`

	rows, err := db.Query(ctx, query, selectedIntervalOption)
	if err != nil {
		return fmt.Errorf("ошибка получения данных из бд для отчёта")
	}

	defer rows.Close()

	var results []string

	results = append(results, "Period, Tickets sold, Daily revenue, Average price, Average price per move")

	for rows.Next() {
		var (
			period       time.Time
			ticketsSold  int
			dailyRevenue float64
			avgPrice     float64
			movingAvg    float64
		)

		err := rows.Scan(
			&period,
			&ticketsSold,
			&dailyRevenue,
			&avgPrice,
			&movingAvg,
		)
		if err != nil {
			return fmt.Errorf("ошибка чтения данных: %w", err)
		}

		rowStr := fmt.Sprintf(
			"%s,%d,%.2f,%.2f,%.2f",
			period.Format("2006-01-02"),
			ticketsSold,
			dailyRevenue,
			avgPrice,
			movingAvg,
		)
		results = append(results, rowStr)
	}

	if err = rows.Err(); err != nil {
		return fmt.Errorf("ошибка при обработке результатов: %w", err)
	}

	if err := r.generateNewReport(results); err != nil {
		return fmt.Errorf("ошибка при создании pdf файла: %s", err)
	}

	return nil
}

func (r *Report) generateNewReport(reportData []string) error {
	if err := os.MkdirAll("./reports", os.ModePerm); err != nil {
		return fmt.Errorf("ошибка при создании директории: %w", err)
	}

	filename := filepath.Join("./reports/report.pdf")

	if _, err := os.Stat(filename); err == nil {
		if err := os.Remove(filename); err != nil {
			return fmt.Errorf("ошибка при удалении существующего файла: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("ошибка при проверке существования файла: %w", err)
	}

	pdf := gofpdf.New("P", "mm", "A4", "")

	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)

	title := "Ticket Sales Report"
	switch r.Interval {
	case 1:
		title += " (week)"
	case 2:
		title += " (month)"
	case 3:
		title += " (year)"
	}
	pdf.Cell(40, 10, title)
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 10)
	colWidths := []float64{40, 30, 30, 30, 40}

	if len(reportData) == 0 {
		return fmt.Errorf("нет данных для отчета")
	}

	headers := strings.Split(reportData[0], ",")
	for i, header := range headers {
		pdf.CellFormat(colWidths[i], 7, header, "1", 0, "C", false, 0, "")
	}
	pdf.Ln(-1)

	for _, row := range reportData[1:] {
		columns := strings.Split(row, ",")
		for i, col := range columns {
			pdf.CellFormat(colWidths[i], 6, col, "1", 0, "C", false, 0, "")
		}
		pdf.Ln(-1)
	}

	if err := pdf.OutputFileAndClose(filename); err != nil {
		return fmt.Errorf("ошибка сохранения PDF: %w", err)
	}

	return nil
}
