package clickhouse

import (
	"context"
	"fmt"
	"hexgonaldb/internal/app/service"
	"hexgonaldb/internal/domain"
	"log"
	"time"

	clickhouse_go "github.com/ClickHouse/clickhouse-go/v2"
)

type Repository struct {
	db clickhouse_go.Conn
}

// NewClickhouseRepository initializes a new connection
func NewClickhouseRepository() *Repository {
	conn, err := clickhouse_go.Open(&clickhouse_go.Options{
		Addr: []string{"localhost:9000"},
		Auth: clickhouse_go.Auth{
			Database: "default",
			Username: "default",
			Password: "",
		},
		Settings: map[string]interface{}{
			"max_execution_time": 300, // increase if large queries run long
			// "max_threads":           8,      // number of threads ClickHouse uses to process query (default = CPU count)
			// "max_insert_block_size": 100000, // how many rows ClickHouse will buffer per insert batch
		},
		Compression: &clickhouse_go.Compression{
			Method: clickhouse_go.CompressionLZ4,
		},
		// Debug: true,
	})
	if err != nil {
		log.Fatalf("failed to connect to ClickHouse: %v", err)
	}

	ctx := context.Background()
	// create table if not exist
	err = conn.Exec(ctx, `
	CREATE TABLE IF NOT EXISTS reports (
		username String,
		username_game String,
		currency String,
		winloss Int64,
		bet Int64,
		turnover Int64,
		payout Float64,
		bet_time DateTime,
		brand_id String,
		brand_name String,
		game_id String,
		game_name String,
		game_type String,
		transaction_id String,
		round_id String
	) ENGINE = MergeTree() ORDER BY (bet_time)
	`)
	if err != nil {
		log.Fatalf("failed to create table: %v", err)
	}

	return &Repository{db: conn}
}

// InsertReport inserts one report record into ClickHouse
func (r *Repository) InsertReport(report domain.Report) error {
	ctx := context.Background()

	query := `
		INSERT INTO reports (
			username,
			username_game,
			currency,
			winloss,
			bet,
			turnover,
			payout,
			bet_time,
			brand_id,
			brand_name,
			game_id,
			game_name,
			game_type,
			transaction_id,
			round_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	err := r.db.Exec(ctx, query,
		report.Username,
		report.UsernameGame,
		report.Currency,
		report.Winloss,
		report.Bet,
		report.Turnover,
		report.Payout,
		report.BetTime,
		report.BrandID,
		report.BrandName,
		report.GameID,
		report.GameName,
		report.GameType,
		report.TransactionID,
		report.RoundID,
	)

	return err
}

func (r *Repository) InsertManyReport(report []domain.Report) error {
	ctx := context.Background()

	query := `
		INSERT INTO reports (
			username,
			username_game,
			currency,
			winloss,
			bet,
			turnover,
			payout,
			bet_time,
			brand_id,
			brand_name,
			game_id,
			game_name,
			game_type,
			transaction_id,
			round_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	for _, report := range report {

		err := r.db.Exec(ctx, query,
			report.Username,
			report.UsernameGame,
			report.Currency,
			report.Winloss,
			report.Bet,
			report.Turnover,
			report.Payout,
			report.BetTime,
			report.BrandID,
			report.BrandName,
			report.GameID,
			report.GameName,
			report.GameType,
			report.TransactionID,
			report.RoundID,
		)

		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Repository) InsertManyReportBatch(report []domain.Report) error {
	ctx := context.Background()

	batch, err := r.db.PrepareBatch(ctx, `
	INSERT INTO reports (
		username,
		username_game,
		currency,
		winloss,
		bet,
		turnover,
		payout,
		bet_time,
		brand_id,
		brand_name,
		game_id,
		game_name,
		game_type,
		transaction_id,
		round_id
		)
	`)
	if err != nil {
		return fmt.Errorf("prepare batch error: %w", err)
	}

	for _, r := range report {
		if err := batch.Append(
			r.Username,
			r.UsernameGame,
			r.Currency,
			r.Winloss,
			r.Bet,
			r.Turnover,
			r.Payout,
			r.BetTime,
			r.BrandID,
			r.BrandName,
			r.GameID,
			r.GameName,
			r.GameType,
			r.TransactionID,
			r.RoundID,
		); err != nil {
			return fmt.Errorf("append batch error: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("send batch error: %w", err)
	}

	return nil
}

func (r *Repository) FindAllReports() (time.Duration, []domain.Report, error) {
	startTime := time.Now()

	ctx := context.Background()
	var reports []domain.Report

	query := "SELECT * FROM reports"
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		errTime := time.Since(startTime)
		return errTime, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var report domain.Report
		if err := rows.Scan(
			&report.Username,
			&report.UsernameGame,
			&report.Currency,
			&report.Winloss,
			&report.Bet,
			&report.Turnover,
			&report.Payout,
			&report.BetTime,
			&report.BrandID,
			&report.BrandName,
			&report.GameID,
			&report.GameName,
			&report.GameType,
			&report.TransactionID,
			&report.RoundID,
		); err != nil {
			errTime := time.Since(startTime)
			return errTime, nil, err
		}
		reports = append(reports, report)
	}

	elapsedTime := time.Since(startTime)

	return elapsedTime, reports, nil
}

func (r *Repository) QueryReport() (time.Duration, []domain.ProfitAggregationResult, error) {

	startTime := time.Now()
	clickQuery := `
		SELECT 
			game_name, 
			SUM(winloss) AS total_profit
		FROM reports
		GROUP BY game_name
		ORDER BY total_profit DESC
	`

	ctx := context.Background()

	rows, err := r.db.Query(ctx, clickQuery)
	if err != nil {
		errTime := time.Since(startTime)
		return errTime, []domain.ProfitAggregationResult{}, fmt.Errorf("ClickHouse query error: %w", err)
	}
	defer rows.Close()

	var allReports []domain.ProfitAggregationResult

	for rows.Next() {
		var r domain.ProfitAggregationResult
		if err := rows.Scan(&r.GameName, &r.TotalProfit); err != nil {
			errTime := time.Since(startTime)
			return errTime, []domain.ProfitAggregationResult{}, fmt.Errorf("ClickHouse scan error: %w", err)
		}

		allReports = append(allReports, r)
	}

	service.TrackResourceUsage("ClickHouse", startTime)
	elapsedTime := time.Since(startTime)

	return elapsedTime, allReports, nil

}

func (r *Repository) QueryReport2() (time.Duration, []domain.SuperAggregationResult, error) {

	startTime := time.Now()
	clickQuery := `
		SELECT 
    		formatDateTime(bet_time, '%Y-%m-%d') AS date,
    		brand_id,
    		game_name,
    		SUM(bet) AS total_bet,
    		SUM(turnover) AS total_turnover,
    		AVG(payout) AS average_payout,
    		COUNT(*) AS total_count,
   			SUM(if(winloss > 0, winloss, 0)) AS positive_win
		FROM reports
		GROUP BY date, brand_id, game_name
		ORDER BY date, brand_id, game_name;
	`
	ctx := context.Background()

	rows, err := r.db.Query(ctx, clickQuery)
	if err != nil {
		errTime := time.Since(startTime)
		return errTime, []domain.SuperAggregationResult{}, fmt.Errorf("ClickHouse query error: %w", err)
	}
	defer rows.Close()

	var allReports []domain.SuperAggregationResult

	for rows.Next() {
		var r domain.SuperAggregationResult
		if err := rows.Scan(&r.Date, &r.BrandID, &r.GameName, &r.TotalBet, &r.TotalTurnover, &r.AveragePayout, &r.TotalCount, &r.PositiveWin); err != nil {
			errTime := time.Since(startTime)
			return errTime, []domain.SuperAggregationResult{}, fmt.Errorf("ClickHouse scan error: %w", err)
		}

		allReports = append(allReports, r)
	}

	service.TrackResourceUsage("ClickHouse", startTime)
	elapsedTime := time.Since(startTime)

	return elapsedTime, allReports, nil

}

func (r *Repository) CountReports() (time.Duration, int64, error) {
	startTime := time.Now()

	ctx := context.Background()

	query := "SELECT COUNT(*) FROM reports"
	var count *uint64
	err := r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		errTime := time.Since(startTime)
		return errTime, 0, err
	}

	convertedCount := int64(*count)

	service.TrackResourceUsage("MongoDB", startTime)
	elapsedTime := time.Since(startTime)

	return elapsedTime, convertedCount, nil
}

func (r *Repository) ClearAll() error {
	ctx := context.Background()

	query := "TRUNCATE TABLE reports"
	err := r.db.Exec(ctx, query)
	if err != nil {
		return err
	}

	return nil
}
