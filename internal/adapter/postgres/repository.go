package postgres

import (
	"fmt"
	"hexgonaldb/internal/app/service"
	"hexgonaldb/internal/domain"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Repository struct {
	db *gorm.DB
}

func NewPostgresRepository() *Repository {
	dsn := "host=localhost user=admin password=secret dbname=app_db port=5432 sslmode=disable"
	db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	db.AutoMigrate(&domain.Report{}) // Auto migrate User
	return &Repository{db}
}

func (r *Repository) CreateReport(report domain.Report) error {
	return r.db.Create(&report).Error
}

func (r *Repository) CreateManyReports(reports []domain.Report) error {
	return r.db.Create(&reports).Error
}

func (r *Repository) FindAllReports() (time.Duration, []domain.Report, error) {
	startTime := time.Now()

	var reports []domain.Report
	err := r.db.Find(&reports).Error

	elapsedTime := time.Since(startTime)

	return elapsedTime, reports, err
}

func (r *Repository) QueryReport() (time.Duration, []domain.ProfitAggregationResult, error) {

	startTime := time.Now()
	var pgResults []domain.ProfitAggregationResult
	err := r.db.Raw(`
		SELECT 
			game_name, 
			SUM(winloss) AS total_profit
		FROM reports
		GROUP BY game_name
		ORDER BY total_profit DESC
	`).Scan(&pgResults).Error
	if err != nil {
		errTime := time.Since(startTime)
		return errTime, []domain.ProfitAggregationResult{}, fmt.Errorf("Postgres query error: %w", err)
	}

	service.TrackResourceUsage("PostgresSQL", startTime)
	elapsedTime := time.Since(startTime)

	return elapsedTime, pgResults, nil
}

func (r *Repository) QueryReport2() (time.Duration, []domain.SuperAggregationResult, error) {

	startTime := time.Now()
	var pgResults []domain.SuperAggregationResult
	err := r.db.Raw(`
		SELECT 
    		TO_CHAR(bet_time, 'YYYY-MM-DD') AS date,
    		brand_id,
    		game_name,
    		SUM(bet) AS total_bet,
    		SUM(turnover) AS total_turnover,
    		AVG(payout) AS average_payout,
    		COUNT(*) AS total_count,
    		SUM(CASE WHEN winloss > 0 THEN winloss ELSE 0 END) AS positive_win
			FROM reports
		GROUP BY date, brand_id, game_name
		ORDER BY date, brand_id, game_name;
	`).Scan(&pgResults).Error
	if err != nil {
		errTime := time.Since(startTime)
		return errTime, []domain.SuperAggregationResult{}, fmt.Errorf("Postgres query error: %w", err)
	}

	service.TrackResourceUsage("PostgresSQL", startTime)
	elapsedTime := time.Since(startTime)

	return elapsedTime, pgResults, nil
}

func (r *Repository) ClearAll() error {
	return r.db.Exec("TRUNCATE TABLE reports").Error
}

func (r *Repository) CountReports() (time.Duration, int64, error) {

	startTime := time.Now()

	query := "SELECT COUNT(*) FROM reports"
	var count int64
	err := r.db.Raw(query).Scan(&count).Error
	if err != nil {
		errTime := time.Since(startTime)
		return errTime, 0, err
	}

	service.TrackResourceUsage("PostgresSQL", startTime)
	elapsedTime := time.Since(startTime)

	return elapsedTime, count, nil
}
