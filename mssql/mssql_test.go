package mssql

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/microsoft/go-mssqldb"
)

func TestRunMssqlPreset(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Start MSSQL container with custom password
	const pw = "MyStr0ngP@ss123!"
	c, err := Run(ctx, WithPassword(pw))
	defer c.Terminate(context.Background())
	if err != nil {
		t.Fatalf("failed to start mssql container: %v", err)
	}

	// Get connection string
	connStr, err := ConnectionString(ctx, c, pw)
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	// Connect to the database
	db, err := sql.Open("sqlserver", connStr)
	if err != nil {
		t.Fatalf("failed to open sql connection: %v", err)
	}
	defer db.Close()

	// Ping the database to ensure it's ready
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("failed to ping sql server: %v", err)
	} else {
		t.Logf("successfully connected to sql server")
	}

	// Execute a simple query
	var one int
	if err := db.QueryRowContext(ctx, "SELECT 1").Scan(&one); err != nil {
		t.Fatalf("failed to query: %v", err)
	} else {
		t.Logf("query result: %d", one)
	}

	// Verify result
	if one != 1 {
		t.Fatalf("unexpected result: got %d, want 1", one)
	}
}
