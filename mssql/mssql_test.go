package mssql

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/microsoft/go-mssqldb"
)

func TestRunMssqlPreset(t *testing.T) {
	// if runtime.GOARCH == "arm64" {
	// 	t.Skip("MSSQL containers are not supported on ARM64 Docker (see microsoft/mssql-docker issues)")
	// }

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	c, err := Run(ctx)
	if err != nil {
		t.Fatalf("failed to start mssql container: %v", err)
	}
	defer c.Terminate(context.Background())

	host, err := c.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get host: %v", err)
	}

	mapped, err := c.MappedPort(ctx, defaultPort)
	if err != nil {
		t.Fatalf("failed to get mapped port: %v", err)
	}

	connStr := fmt.Sprintf(
		"sqlserver://%s:%s@%s:%s?database=master&encrypt=disable",
		DefaultUser,
		DefaultPassword,
		host,
		mapped.Port(),
	)

	db, err := sql.Open("sqlserver", connStr)
	if err != nil {
		t.Fatalf("failed to open sql connection: %v", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("failed to ping sql server: %v", err)
	}

	var one int
	if err := db.QueryRowContext(ctx, "SELECT 1").Scan(&one); err != nil {
		t.Fatalf("failed to query: %v", err)
	}
	if one != 1 {
		t.Fatalf("unexpected result: got %d, want 1", one)
	}
}
