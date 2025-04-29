package main

import (
	"fmt"
	"hexgonaldb/internal/adapter/clickhouse"
	"hexgonaldb/internal/adapter/mongo"
	"hexgonaldb/internal/adapter/postgres"
	"hexgonaldb/internal/app/service"
	"hexgonaldb/internal/domain"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

const (
	totalReports  = 5_000_000 // total reports to generate
	batchSize     = 1000      // how many reports in each batch
	maxGoroutines = 50        // how many goroutines at the same time
)

func main() {

	// runtime.GOMAXPROCS(runtime.NumCPU())

	// Init Database Adapters
	fmt.Println("Initializing database adapters...")

	pgRepo := postgres.NewPostgresRepository()
	fmt.Printf("[Postgres] connected to PostgreSQL database\n")

	mongoRepo := mongo.NewMongoRepository()
	fmt.Printf("[MongoDB] connected to MongoDB database\n")

	chRepo := clickhouse.NewClickhouseRepository()
	fmt.Printf("[ClickHouse] connected to ClickHouse database\n\n")

	// pgRepo.ClearAll()
	// mongoRepo.ClearAll("reports")
	// chRepo.ClearAll()

	// Init Service
	appService := service.NewService(pgRepo, mongoRepo, chRepo)

	start := time.Now()

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxGoroutines) // Semaphore channel

	batchGenerateReports := totalReports / batchSize

	skipInsert := true

	currentReport := totalReports
	for i := 0; i < batchGenerateReports; i++ {
		if skipInsert {
			fmt.Println("Skipping insert for testing...")
			break
		}

		// Generate reports in batches
		allReports := appService.GenerateReports(batchSize)

		fmt.Printf("Generating reports... round %d/%d \n", i+1, batchGenerateReports)

		// Split into batches
		for i := 0; i < len(allReports); i += batchSize {
			end := i + batchSize
			if end > len(allReports) {
				end = len(allReports)
			}
			batch := allReports[i:end]

			wg.Add(1)
			semaphore <- struct{}{} // acquire slot

			go func(batchReports []domain.Report) {
				defer wg.Done()
				defer func() { <-semaphore }() // release slot

				startTime := time.Now()

				startTimePostgres := time.Now()
				// PostgreSQL batch insert
				if err := pgRepo.CreateManyReports(batchReports); err != nil {
					log.Printf("[Postgres] batch insert error: %v\n", err)
				} else {
					fmt.Println("[Postgres] batch insert success took:", time.Since(startTimePostgres))
				}

				var document []any
				for _, report := range batchReports {
					document = append(document, report)
				}

				startTimeMongo := time.Now()
				// MongoDB batch insert
				if err := mongoRepo.CreateManyDocuments("reports", document); err != nil {
					log.Printf("[MongoDB] batch insert error: %v\n", err)
				} else {
					fmt.Println("[MongoDB] batch insert success took:", time.Since(startTimeMongo))
				}

				startTimeClickHouse := time.Now()
				// ClickHouse batch insert
				if err := chRepo.InsertManyReportBatch(batchReports); err != nil {
					log.Printf("[ClickHouse] batch insert error: %v\n", err)
				} else {
					fmt.Println("[ClickHouse] batch insert success took:", time.Since(startTimeClickHouse))
				}

				currentReport -= len(batchReports)
				fmt.Println("left reports left to insert:", currentReport, " percent complete:", 100-(currentReport*100)/totalReports, " %")
				fmt.Println("All Batch insert took:", time.Since(startTime))

			}(batch)
		}
	}

	wg.Wait()

	// // Measure Read Performance
	// fmt.Println("\nReading from all databases...")

	// filter := map[string]interface{}{}
	// mongoTime, mongoReports, err := mongoRepo.FindManyDocuments("reports", filter)
	// if err != nil {
	// 	log.Printf("Error finding reports in MongoDB: %v\n", err)
	// }

	// fmt.Println("need more : ", totalReports-len(mongoReports))

	// postgresTime, postgresReports, err := pgRepo.FindAllReports()
	// if err != nil {
	// 	log.Printf("Error finding reports in PostgreSQL: %v\n", err)
	// }
	// clickhouseTime, clickhouseReports, err := chRepo.FindAllReports()
	// if err != nil {
	// 	log.Printf("Error finding reports in ClickHouse: %v\n", err)
	// }

	// fmt.Println("[QueryAll]")
	// fmt.Println("MongoDB Time:", mongoTime, "Found:", len(mongoReports))
	// fmt.Println("PostgreSQL Time:", postgresTime, "Found:", len(postgresReports))
	// fmt.Println("Clickhouse Time:", clickhouseTime, "Found:", len(clickhouseReports))
	// fmt.Println("---------------------")

	fmt.Println("----- CountDocuments -----")
	filterCount := bson.M{}
	mongoTime, mongoCount, err := mongoRepo.CountDocuments("reports", filterCount)
	if err != nil {
		log.Printf("Error counting documents in MongoDB: %v\n", err)
	}

	postgresTime, postgresCount, err := pgRepo.CountReports()
	if err != nil {
		log.Printf("Error counting reports in PostgreSQL: %v\n", err)
	}

	clickhouseTime, clickhouseCount, err := chRepo.CountReports()
	if err != nil {
		log.Printf("Error counting reports in ClickHouse: %v\n", err)
	}

	fmt.Println("Summary:")
	fmt.Printf("[MongoDB] Time: %.2f seconds, Found: %d\n", mongoTime.Seconds(), mongoCount)
	fmt.Printf("[PostgreSQL] Time: %.2f seconds, Found: %d\n", postgresTime.Seconds(), postgresCount)
	fmt.Printf("[ClickHouse] Time: %.2f seconds, Found: %d\n", clickhouseTime.Seconds(), clickhouseCount)
	fmt.Println("---------------------")
	fmt.Println("")

	fmt.Println("----- Simple Aggregation -----")
	mongoTime, mongoAggregation, err := mongoRepo.AggregationReports("reports")
	if err != nil {
		log.Printf("Error aggregating reports in MongoDB: %v\n", err)
	} else {
		var sum int64
		for _, result := range mongoAggregation {
			sum += result.TotalProfit
		}

		fmt.Println("result sum in MongoDb: ", sum)
	}

	postgresTime, postgresAggregation, err := pgRepo.QueryReport()
	if err != nil {
		log.Printf("Error aggregating reports in PostgreSQL: %v\n", err)
	} else {
		var sum int64
		for _, result := range postgresAggregation {
			sum += result.TotalProfit
		}

		fmt.Println("result sum in PostgreSQL:", sum)
	}

	clickhouseTime, clickhouseAggregation, err := chRepo.QueryReport()
	if err != nil {
		log.Printf("Error aggregating reports in ClickHouse: %v\n", err)
	} else {
		var sum int64
		for _, result := range clickhouseAggregation {
			sum += result.TotalProfit
		}
		fmt.Println("result sum in ClickHouse:", sum)
	}

	fmt.Println("Summary:")
	fmt.Printf("[MongoDB] Time: %.2f seconds, Found: %d\n", mongoTime.Seconds(), len(mongoAggregation))
	fmt.Printf("[PostgreSQL] Time: %.2f seconds, Found: %d\n", postgresTime.Seconds(), len(postgresAggregation))
	fmt.Printf("[ClickHouse] Time: %.2f seconds, Found: %d\n", clickhouseTime.Seconds(), len(clickhouseAggregation))
	fmt.Println("---------------------")
	fmt.Println("")

	fmt.Println("----- Complex Aggregation2 -----")
	mongoTime2, mongoAggregation2, err := mongoRepo.AggregationReports2("reports")
	if err != nil {
		log.Printf("Error aggregating reports in MongoDB: %v\n", err)
	} else {
		// minDate, maxDate, totalDays, err := service.FindDateRangeMongo(mongoAggregation2)
		// if err != nil {
		// 	log.Fatal(err)
		// }

		// fmt.Printf("MongoDB aggregation took: %.2f seconds, Found: %d, Date Range: %s to %s, Total Days: %d\n", mongoTime2.Seconds(), len(mongoAggregation2), minDate, maxDate, totalDays)
	}

	postgresTime2, postgresAggregation2, err := pgRepo.QueryReport2()
	if err != nil {
		log.Printf("Error aggregating reports in PostgreSQL: %v\n", err)
	} else {
		// minDate, maxDate, totalDays, err := service.FindDateRange(postgresAggregation2)
		// if err != nil {
		// 	log.Fatal(err)
		// }

		// fmt.Printf("PostgreSQL aggregation took: %.2f seconds, Found: %d, Date Range: %s to %s, Total Days: %d\n", postgresTime2.Seconds(), len(postgresAggregation2), minDate, maxDate, totalDays)
	}

	clickhouseTime2, clickhouseAggregation2, err := chRepo.QueryReport2()
	if err != nil {
		log.Printf("Error aggregating reports in ClickHouse: %v\n", err)
	} else {
		// minDate, maxDate, totalDays, err := service.FindDateRange(clickhouseAggregation2)
		// if err != nil {
		// 	log.Fatal(err)
		// }

		// fmt.Printf("ClickHouse aggregation took: %.2f seconds, Found: %d, Date Range: %s to %s, Total Days: %d\n", clickhouseTime2.Seconds(), len(clickhouseAggregation2), minDate, maxDate, totalDays)
	}

	fmt.Println("Summary:")
	fmt.Printf("[MongoDB] Time: %.2f seconds, Found: %d\n", mongoTime2.Seconds(), len(mongoAggregation2))
	fmt.Printf("[PostgreSQL] Time: %.2f seconds, Found: %d\n", postgresTime2.Seconds(), len(postgresAggregation2))
	fmt.Printf("[ClickHouse] Time: %.2f seconds, Found: %d\n", clickhouseTime2.Seconds(), len(clickhouseAggregation2))
	fmt.Println("---------------------")
	fmt.Println("")

	fmt.Println("Done. Total Time:", time.Since(start))
}
