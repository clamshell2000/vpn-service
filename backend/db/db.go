package db

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/vpn-service/backend/src/config"
	"github.com/vpn-service/backend/src/utils"
)

var (
	// DB is the database connection
	DB *sqlx.DB
)

// Connect connects to the database
func Connect(cfg *config.Config) error {
	// Build connection string
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
	)

	// Connect to database
	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Ping database to verify connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	// Set global DB variable
	DB = db

	utils.LogInfo("Connected to database")
	return nil
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// RunMigrations runs database migrations
func RunMigrations(cfg *config.Config) error {
	// TODO: Implement database migrations
	// This would typically use a migration library like golang-migrate
	utils.LogInfo("Running database migrations")
	return nil
}
