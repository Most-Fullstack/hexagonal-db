package app

import (
	"hexgonaldb/internal/domain"
)

// Define interfaces that adapter must implement (Ports)

type PostgresRepository interface {
	CreateReport(report domain.Report) error
}

type MongoRepository interface {
	CreateOneDocument(collection string, document interface{}) error
	CreateManyDocuments(collection string, documents []interface{}) error
}

type ClickhouseRepository interface {
}
