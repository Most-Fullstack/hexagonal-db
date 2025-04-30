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

