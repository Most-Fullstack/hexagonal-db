package mongo

import (
	"context"
	"hexgonaldb/internal/app/service"
	"hexgonaldb/internal/domain"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Repository struct {
	client *mongo.Client
}

func NewMongoRepository() *Repository {
	client, _ := mongo.NewClient(options.Client().ApplyURI("mongodb://admin:secret@localhost:27017"))
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client.Connect(ctx)
	return &Repository{client}
}

func (r *Repository) CreateOneDocument(collection string, document interface{}) error {
	collectionRef := r.client.Database("app_db").Collection(collection)
	_, err := collectionRef.InsertOne(context.Background(), document)
	return err
}

func (r *Repository) CreateManyDocuments(collection string, documents []any) error {
	collectionRef := r.client.Database("app_db").Collection(collection)
	_, err := collectionRef.InsertMany(context.Background(), documents)
	return err
}

func (r *Repository) FindManyDocuments(collection string, filter interface{}) (time.Duration, []interface{}, error) {
	startTime := time.Now()

	collectionRef := r.client.Database("app_db").Collection(collection)
	cursor, err := collectionRef.Find(context.Background(), filter)
	if err != nil {
		errTime := time.Since(startTime)
		return errTime, nil, err
	}
	defer cursor.Close(context.Background())

	var results []interface{}
	for cursor.Next(context.Background()) {
		var result interface{}
		if err := cursor.Decode(&result); err != nil {
			errTime := time.Since(startTime)

			return errTime, nil, err
		}
		results = append(results, result)
	}

	service.TrackResourceUsage("MongoDB", startTime)
	elapsedTime := time.Since(startTime)

	return elapsedTime, results, nil
}

func (r *Repository) CountDocuments(collection string, filter interface{}) (time.Duration, int64, error) {
	startTime := time.Now()
	collectionRef := r.client.Database("app_db").Collection(collection)
	count, err := collectionRef.CountDocuments(context.Background(), filter)
	if err != nil {
		errTime := time.Since(startTime)
		return errTime, 0, err
	}

	service.TrackResourceUsage("MongoDB", startTime)
	elapsedTime := time.Since(startTime)

	return elapsedTime, count, err
}

func (r *Repository) AggregationReports(collection string) (time.Duration, []domain.ProfitAggregationResult, error) {
	startTime := time.Now()

	mongoPipeline := mongo.Pipeline{
		{{
			Key: "$group",
			Value: bson.D{
				{Key: "_id", Value: "$game_name"},
				{Key: "total_profit", Value: bson.D{{Key: "$sum", Value: "$winloss"}}},
			},
		}},
		{{
			Key: "$sort",
			Value: bson.D{
				{Key: "total_profit", Value: -1}, // sort descending by profit (optional)
			},
		}},
	}

	collectionRef := r.client.Database("app_db").Collection(collection)
	cursor, err := collectionRef.Aggregate(context.Background(), mongoPipeline)
	if err != nil {
		return time.Since(startTime), nil, err
	}
	defer cursor.Close(context.Background())

	var tempResults []domain.ProfitAggregationResult
	if err := cursor.All(context.Background(), &tempResults); err != nil {
		return time.Since(startTime), nil, err
	}

	service.TrackResourceUsage("MongoDB", startTime)
	elapsedTime := time.Since(startTime)

	return elapsedTime, tempResults, nil
}

func (r *Repository) AggregationReports2(collection string) (time.Duration, []domain.MongoAggregationResult, error) {
	startTime := time.Now()

	mongoPipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{
				{Key: "date", Value: bson.D{
					{Key: "$dateToString", Value: bson.D{
						{Key: "format", Value: "%Y-%m-%d"},
						{Key: "date", Value: "$bet_time"},
					}},
				}},
				{Key: "brand_id", Value: "$brand_id"},
				{Key: "game_name", Value: "$game_name"},
			}},
			{Key: "total_bet", Value: bson.D{{Key: "$sum", Value: "$bet"}}},
			{Key: "total_turnover", Value: bson.D{{Key: "$sum", Value: "$turnover"}}},
			{Key: "average_payout", Value: bson.D{{Key: "$avg", Value: "$payout"}}},
			{Key: "total_count", Value: bson.D{{Key: "$sum", Value: 1}}},
			{Key: "positive_win", Value: bson.D{
				{Key: "$sum", Value: bson.D{
					{Key: "$cond", Value: bson.A{
						bson.D{{Key: "$gt", Value: bson.A{"$winloss", 0}}},
						"$winloss",
						0,
					}},
				}},
			}},
		}}},
		{{Key: "$sort", Value: bson.D{
			{Key: "_id.date", Value: 1},
			{Key: "_id.brand_id", Value: 1},
			{Key: "_id.game_name", Value: 1},
		}}},
	}

	collectionRef := r.client.Database("app_db").Collection(collection)
	cursor, err := collectionRef.Aggregate(context.Background(), mongoPipeline)
	if err != nil {
		return time.Since(startTime), nil, err
	}
	defer cursor.Close(context.Background())

	var tempResults []domain.MongoAggregationResult
	if err := cursor.All(context.Background(), &tempResults); err != nil {
		return time.Since(startTime), nil, err
	}

	service.TrackResourceUsage("MongoDB", startTime)
	elapsedTime := time.Since(startTime)

	return elapsedTime, tempResults, nil
}

func (r *Repository) ClearAll(collection string) error {
	ctx := context.Background()

	collectionRef := r.client.Database("app_db").Collection(collection)
	_, err := collectionRef.DeleteMany(ctx, bson.D{})
	return err
}
