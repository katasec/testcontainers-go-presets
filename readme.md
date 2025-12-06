# 📦 testcontainers-go-presets

A lightweight collection of **convenience presets for Testcontainers-Go**, inspired by **[Testcontainers for .NET](
https://github.com/testcontainers/testcontainers-dotnet)**

The goal is to reduce boiler plate with predictable defaults for spinning up containers.This library wraps `github.com/testcontainers/testcontainers-go` with small, composable helpers that make tests easier to write and easier to read.


## MS SQL Example

### Simple Launch

Uses a default password `Passw0rd123`:

```go
import "github.com/katasec/testcontainers-go-presets/mssql"

ctx := context.Background()

c, err := mssql.Run(ctx)
if err != nil {
    panic(err)
}
defer c.Terminate(context.Background())
```

### Custom Password

```go
c, err := mssql.Run(ctx, mssql.WithPassword("MyStr0ngP@ss"))
```

### Sample Test

```go
func TestMssql(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), time.Minute*2)
    defer cancel()

    // Spin up MS SQL container
    c, err := mssql.Run(ctx, mssql.WithPassword("TestP@ss123"))
    if err != nil {
        t.Fatalf("failed to start: %v", err)
    }
    t.Cleanup(func() { c.Terminate(context.Background()) })

    // Get the connection string
    connStr, err := mssql.ConnectionString(ctx, c, "TestP@ss123")
    if err != nil {
        t.Fatalf("failed to build connection string: %v", err)
    }

    // Connect to server
    db, err := sql.Open("sqlserver", connStr)
    if err != nil {
        t.Fatalf("failed to open: %v", err)
    }
    defer db.Close()

    // Ping the DB
    if err := db.PingContext(ctx); err != nil {
        t.Fatalf("failed to ping: %v", err)
    }
}
```
