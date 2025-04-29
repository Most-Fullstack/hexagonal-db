package service

import (
	"fmt"
	"hexgonaldb/internal/app"
	"hexgonaldb/internal/domain"
	"math/rand"
	"runtime"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	postgres app.PostgresRepository
	mongo    app.MongoRepository
	click    app.ClickhouseRepository
}

func NewService(pg app.PostgresRepository, mongo app.MongoRepository, click app.ClickhouseRepository) *Service {
	return &Service{pg, mongo, click}
}

// func (s *Service) GenerateReports(count int) []domain.Report {

// 	fmt.Println("Generating reports...")

// 	var reports []domain.Report
// 	now := time.Now()

// 	game := []string{"pgsoft", "evolution", "evolutionlive", "netent", "playtech", "pragmatic", "redtiger", "quickspin", "microgaming", "yggdrasil"}

// 	completePercent := 0

// 	for i := 0; i < count; i++ {

// 		username := uuid.NewString()
// 		usernameGame := fmt.Sprintf("%s_%s", username, game[rand.Intn(len(game))])

// 		r := domain.Report{
// 			Username:      username,
// 			UsernameGame:  usernameGame,
// 			Currency:      "USD",
// 			Winloss:       rand.Int63n(10000) - 5000, // -5000 to +5000
// 			Bet:           rand.Int63n(10000),
// 			Turnover:      rand.Int63n(20000),
// 			Payout:        rand.Float64() * 100,
// 			BetTime:       now.Add(-time.Duration(rand.Intn(100000)) * time.Second),
// 			BrandID:       fmt.Sprintf("brand%d", rand.Intn(10)),
// 			BrandName:     fmt.Sprintf("Brand %d", rand.Intn(10)),
// 			GameID:        fmt.Sprintf("game%d", rand.Intn(100)),
// 			GameName:      fmt.Sprintf("Game %d", rand.Intn(100)),
// 			GameType:      fmt.Sprintf("type%d", rand.Intn(5)),
// 			TransactionID: fmt.Sprintf("tx%d", rand.Int63()),
// 			RoundID:       fmt.Sprintf("round%d", rand.Int63()),
// 		}
// 		reports = append(reports, r)

// 		completePercent = int((float64(i) / float64(count)) * 100)
// 		if completePercent%2 == 0 {
// 			fmt.Printf("\rGenerating reports... %d%% complete", completePercent)
// 		}
// 	}

// 	return reports
// }

func (s *Service) GenerateReports(count int) []domain.Report {
	var (
		reports = make([]domain.Report, count)
		now     = time.Now()
		game    = []string{"pgsoft", "evolution", "evolutionlive", "netent", "playtech", "pragmatic", "redtiger", "quickspin", "microgaming", "yggdrasil"}
	)

	for i := 0; i < count; i++ {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))

		username := uuid.NewString()
		usernameGame := fmt.Sprintf("%s_%s", username, game[r.Intn(len(game))])

		reports[i] = domain.Report{
			Username:      username,
			UsernameGame:  usernameGame,
			Currency:      "USD",
			Winloss:       r.Int63n(10000) - 5000,
			Bet:           r.Int63n(10000),
			Turnover:      r.Int63n(20000),
			Payout:        r.Float64() * 100,
			BetTime:       now.Add(-time.Duration(r.Intn(100000)) * time.Minute),
			BrandID:       fmt.Sprintf("brand%d", r.Intn(10)),
			BrandName:     fmt.Sprintf("Brand %d", r.Intn(10)),
			GameID:        fmt.Sprintf("game%d", r.Intn(100)),
			GameName:      fmt.Sprintf("Game %d", r.Intn(100)),
			GameType:      fmt.Sprintf("type%d", r.Intn(5)),
			TransactionID: fmt.Sprintf("tx%d", r.Int63()),
			RoundID:       fmt.Sprintf("round%d", r.Int63()),
		}

	}
	return reports
}

func FindDateRangeMongo(results []domain.MongoAggregationResult) (minDateStr, maxDateStr string, totalDays int, err error) {
	if len(results) == 0 {
		return "", "", 0, fmt.Errorf("no data to find range")
	}

	const layout = "2006-01-02" // your date format is like "YYYY-MM-DD"

	minDate, err := time.Parse(layout, results[0].ID.Date)
	if err != nil {
		return "", "", 0, fmt.Errorf("invalid date format: %w", err)
	}
	maxDate := minDate

	for _, r := range results {
		d, err := time.Parse(layout, r.ID.Date)
		if err != nil {
			return "", "", 0, fmt.Errorf("invalid date format: %w", err)
		}
		if d.Before(minDate) {
			minDate = d
		}
		if d.After(maxDate) {
			maxDate = d
		}
	}

	duration := maxDate.Sub(minDate)
	totalDays = int(duration.Hours()/24) + 1 // +1 to include both start and end day

	return minDate.Format(layout), maxDate.Format(layout), totalDays, nil
}

func FindDateRange(results []domain.SuperAggregationResult) (minDateStr, maxDateStr string, totalDays int, err error) {
	if len(results) == 0 {
		return "", "", 0, fmt.Errorf("no data to find range")
	}

	const layout = "2006-01-02" // your date format is like "YYYY-MM-DD"

	minDate, err := time.Parse(layout, results[0].Date)
	if err != nil {
		return "", "", 0, fmt.Errorf("invalid date format: %w", err)
	}
	maxDate := minDate

	for _, r := range results {
		d, err := time.Parse(layout, r.Date)
		if err != nil {
			return "", "", 0, fmt.Errorf("invalid date format: %w", err)
		}
		if d.Before(minDate) {
			minDate = d
		}
		if d.After(maxDate) {
			maxDate = d
		}
	}

	duration := maxDate.Sub(minDate)
	totalDays = int(duration.Hours()/24) + 1 // +1 to include both start and end day

	return minDate.Format(layout), maxDate.Format(layout), totalDays, nil
}

func TrackResourceUsage(operation string, start time.Time) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	duration := time.Since(start)

	fmt.Printf("[%s] Done in %.2f | Memory Used: %.2f MB | Alloc: %.2f MB | TotalAlloc: %.2f MB\n",
		operation,
		duration.Seconds(),
		float64(mem.Sys)/1024/1024,
		float64(mem.Alloc)/1024/1024,
		float64(mem.TotalAlloc)/1024/1024,
	)
}
