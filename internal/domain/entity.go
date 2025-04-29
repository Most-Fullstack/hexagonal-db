package domain

import "time"

type Report struct {
	Username      string    `json:"username" bson:"username"`
	UsernameGame  string    `json:"username_game" bson:"username_game"`
	Currency      string    `json:"currency" bson:"currency"`
	Winloss       int64     `json:"winloss" bson:"winloss"`
	Bet           int64     `json:"bet" bson:"bet"`
	Turnover      int64     `json:"turnover" bson:"turnover"`
	Payout        float64   `json:"payout" bson:"payout"`
	BetTime       time.Time `json:"bet_time" bson:"bet_time"`
	BrandID       string    `json:"brand_id" bson:"brand_id"`
	BrandName     string    `json:"brand_name" bson:"brand_name"`
	GameID        string    `json:"game_id" bson:"game_id"`
	GameName      string    `json:"game_name" bson:"game_name"`
	GameType      string    `json:"game_type" bson:"game_type"`
	TransactionID string    `json:"transaction_id" bson:"transaction_id"`
	RoundID       string    `json:"round_id" bson:"round_id"`
}

type AggregationResult struct {
	Date     string `json:"date"`
	GameName string `json:"game_name"`
	TotalBet int64  `json:"total_bet"`
}

type SuperAggregationResult struct {
	Date          string  `json:"date"`
	BrandID       string  `json:"brand_id"`
	GameName      string  `json:"game_name"`
	TotalBet      int64   `json:"total_bet"`
	TotalTurnover int64   `json:"total_turnover"`
	AveragePayout float64 `json:"average_payout"`
	TotalCount    uint64  `json:"total_count"`
	PositiveWin   int64   `json:"positive_win"`
}

type MongoAggregationResult struct {
	ID struct {
		Date     string `bson:"date"`
		BrandID  string `bson:"brand_id"`
		GameName string `bson:"game_name"`
	} `bson:"_id"`
	TotalBet      int64   `bson:"total_bet"`
	TotalTurnover int64   `bson:"total_turnover"`
	AveragePayout float64 `bson:"average_payout"`
	TotalCount    int64   `bson:"total_count"`
	PositiveWin   int64   `bson:"positive_win"`
}

type ProfitAggregationResult struct {
	GameName    string `json:"game_name" bson:"game_name"`
	TotalProfit int64  `json:"total_profit" bson:"total_profit"`
}
