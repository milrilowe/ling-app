package testutil

import (
	"fmt"
	"os"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"ling-app/api/internal/db"
	"ling-app/api/internal/models"
)

// TestDB wraps a database connection for integration tests.
// It provides helpers for setup and cleanup.
type TestDB struct {
	*db.DB
	t *testing.T
}

// NewTestDB creates a new test database connection.
// It uses TEST_DATABASE_URL env var, falling back to a default test database.
// Returns nil if no test database is available (skips test).
func NewTestDB(t *testing.T) *TestDB {
	t.Helper()

	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		// Default to local test database
		dbURL = "postgresql://lingapp:lingapp@localhost:5432/lingapp_test"
	}

	// Try to connect
	gormDB, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Skipf("Skipping integration test: database not available: %v", err)
		return nil
	}

	testDB := &TestDB{
		DB: &db.DB{DB: gormDB},
		t:  t,
	}

	// Run migrations
	if err := testDB.RunMigrations(
		&models.User{},
		&models.Session{},
		&models.Thread{},
		&models.Message{},
		&models.Subscription{},
		&models.Credits{},
		&models.CreditTransaction{},
		&models.PhonemeStats{},
		&models.PhonemeSubstitution{},
	); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return testDB
}

// Cleanup removes all data from test tables.
// Call this in t.Cleanup() or defer.
func (tdb *TestDB) Cleanup() {
	if tdb == nil {
		return
	}

	// Delete in reverse order of foreign key dependencies
	tables := []string{
		"phoneme_substitutions",
		"phoneme_stats",
		"credit_transactions",
		"credits",
		"subscriptions",
		"messages",
		"threads",
		"sessions",
		"users",
	}

	for _, table := range tables {
		tdb.Exec(fmt.Sprintf("DELETE FROM %s", table))
	}
}

// TruncateTables truncates all tables (faster than DELETE for large datasets).
func (tdb *TestDB) TruncateTables() {
	if tdb == nil {
		return
	}

	tables := []string{
		"phoneme_substitutions",
		"phoneme_stats",
		"credit_transactions",
		"credits",
		"subscriptions",
		"messages",
		"threads",
		"sessions",
		"users",
	}

	// Disable FK checks for truncate
	tdb.Exec("SET session_replication_role = 'replica'")
	for _, table := range tables {
		tdb.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
	}
	tdb.Exec("SET session_replication_role = 'origin'")
}
