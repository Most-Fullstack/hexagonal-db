# 📊 Go-Gin Database Benchmarking Project (Hexagonal Architecture)

This project is a Golang web service built with the [Gin](https://github.com/gin-gonic/gin) framework and structured using **Hexagonal Architecture** (Ports & Adapters).  
It benchmarks the performance of three different databases:

- 🏎️ [ClickHouse](https://clickhouse.com/)
- 🍃 [MongoDB](https://www.mongodb.com/)
- 🐘 [PostgreSQL](https://www.postgresql.org/)

Each database is implemented as an **adapter**, allowing for clean separation between core business logic and infrastructure.

---

## 🧱 Architecture

This project follows the **Hexagonal (Clean) Architecture**:

## 🐳 Docker Setup

The project includes a `docker-compose.yml` file to spin up all three databases quickly:

```bash
docker-compose up -d
```

## Runing command
```bash
go run cmd/server/main.go
```

make sure runing docker compose before


## Results

### Count Documents
- **MongoDB**: 39.47 seconds | Memory: 11.33 MB | Alloc: 2.20 MB | TotalAlloc: 2.20 MB | Found: 15,000,000
- **PostgreSQL**: 18.89 seconds | Memory: 11.33 MB | Alloc: 2.24 MB | TotalAlloc: 2.24 MB | Found: 15,000,000
- **ClickHouse**: 0.01 seconds | Memory: 11.33 MB | Alloc: 2.26 MB | TotalAlloc: 2.26 MB | Found: 15,000,000

### Simple Aggregation
- **MongoDB**: 28.08 seconds | Memory: 11.33 MB | Alloc: 2.35 MB | TotalAlloc: 2.35 MB | Result Sum: -16,739,605 | Found: 100
- **PostgreSQL**: 18.43 seconds | Memory: 11.33 MB | Alloc: 2.40 MB | TotalAlloc: 2.40 MB | Result Sum: -16,739,605 | Found: 100
- **ClickHouse**: 0.47 seconds | Memory: 11.33 MB | Alloc: 2.54 MB | TotalAlloc: 2.54 MB | Result Sum: -16,739,605 | Found: 100

### Complex Aggregation
- **MongoDB**: 77.81 seconds | Memory: 64.15 MB | Alloc: 27.25 MB | TotalAlloc: 93.21 MB | Found: 71,000
- **PostgreSQL**: 76.26 seconds | Memory: 64.27 MB | Alloc: 18.50 MB | TotalAlloc: 150.72 MB | Found: 71,000
- **ClickHouse**: 2.25 seconds | Memory: 64.27 MB | Alloc: 33.34 MB | TotalAlloc: 211.09 MB | Found: 71,000