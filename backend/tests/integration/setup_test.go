package integration

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// No DB available — tests will skip
		os.Exit(m.Run())
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		panic("failed to connect to test database: " + err.Error())
	}
	testPool = pool

	code := m.Run()

	// Clean up only test-created data (scoped to test email domain)
	if _, err := testPool.Exec(ctx, "DELETE FROM tasks WHERE created_by IN (SELECT id FROM users WHERE email LIKE '%@test.com')"); err != nil {
		log.Printf("warning: failed to clean up test tasks: %v", err)
	}
	if _, err := testPool.Exec(ctx, "DELETE FROM projects WHERE owner_id IN (SELECT id FROM users WHERE email LIKE '%@test.com')"); err != nil {
		log.Printf("warning: failed to clean up test projects: %v", err)
	}
	if _, err := testPool.Exec(ctx, "DELETE FROM users WHERE email LIKE '%@test.com'"); err != nil {
		log.Printf("warning: failed to clean up test users: %v", err)
	}
	testPool.Close()

	os.Exit(code)
}
