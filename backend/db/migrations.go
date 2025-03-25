package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/vpn-service/backend/src/config"
	"github.com/vpn-service/backend/src/utils"
)

// MigrationManager manages database migrations
type MigrationManager struct {
	config *config.Config
	db     *sql.DB
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(cfg *config.Config, db *sql.DB) *MigrationManager {
	return &MigrationManager{
		config: cfg,
		db:     db,
	}
}

// RunMigrations runs all pending migrations
func (mm *MigrationManager) RunMigrations() error {
	// Get migrations directory
	migrationsDir := filepath.Join("db", "migrations")
	absPath, err := filepath.Abs(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for migrations directory: %v", err)
	}

	// Check if migrations directory exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("migrations directory does not exist: %s", absPath)
	}

	// Create postgres driver
	driver, err := postgres.WithInstance(mm.db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %v", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", absPath),
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %v", err)
	}

	// Run migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %v", err)
	}

	utils.LogInfo("Database migrations completed successfully")
	return nil
}

// GetMigrationVersion gets the current migration version
func (mm *MigrationManager) GetMigrationVersion() (uint, bool, error) {
	// Get migrations directory
	migrationsDir := filepath.Join("db", "migrations")
	absPath, err := filepath.Abs(migrationsDir)
	if err != nil {
		return 0, false, fmt.Errorf("failed to get absolute path for migrations directory: %v", err)
	}

	// Create postgres driver
	driver, err := postgres.WithInstance(mm.db, &postgres.Config{})
	if err != nil {
		return 0, false, fmt.Errorf("failed to create postgres driver: %v", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", absPath),
		"postgres",
		driver,
	)
	if err != nil {
		return 0, false, fmt.Errorf("failed to create migrate instance: %v", err)
	}

	// Get version
	version, dirty, err := m.Version()
	if err == migrate.ErrNilVersion {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("failed to get migration version: %v", err)
	}

	return version, dirty, nil
}

// MigrateDown rolls back the last migration
func (mm *MigrationManager) MigrateDown() error {
	// Get migrations directory
	migrationsDir := filepath.Join("db", "migrations")
	absPath, err := filepath.Abs(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for migrations directory: %v", err)
	}

	// Create postgres driver
	driver, err := postgres.WithInstance(mm.db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %v", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", absPath),
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %v", err)
	}

	// Run down migration
	if err := m.Steps(-1); err != nil {
		return fmt.Errorf("failed to run down migration: %v", err)
	}

	utils.LogInfo("Down migration completed successfully")
	return nil
}
