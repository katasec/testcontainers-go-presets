package mssql

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	DefaultImage    = "mcr.microsoft.com/mssql/server:2022-latest"
	DefaultUser     = "sa"
	DefaultPassword = "Passw0rd123" // for tests only
	defaultPort     = "1433/tcp"
)

// RunOptions holds optional flags and overrides for Run.
type RunOptions struct {
	DisableReaper bool
	Password      string // optional override for SA password
}

// RunOption configures RunOptions.
type RunOption func(*RunOptions)

// WithReaperDisabled disables Ryuk via ~/.testcontainers.properties
// so that containers are not auto-removed after the test process exits.
func WithReaperDisabled() RunOption {
	return func(o *RunOptions) {
		o.DisableReaper = true
	}
}

// WithPassword overrides the SA password used by SQL Server.
func WithPassword(pw string) RunOption {
	return func(o *RunOptions) {
		o.Password = pw
	}
}

// Run starts an MSSQL container with optional RunOptions.
//
// Example:
//
//	c, err := mssql.Run(ctx,
//	    mssql.WithPassword("MyStr0ngP@ss"),
//	    mssql.WithReaperDisabled(),
//	)
func Run(ctx context.Context, opts ...RunOption) (testcontainers.Container, error) {

	// Create empty options
	options := RunOptions{}

	// Apply provided options by calling each funct
	for _, optionFunc := range opts {
		optionFunc(&options)
	}

	// Disable reaper if requested
	if options.DisableReaper {
		if err := disableReaperFile(); err != nil {
			return nil, fmt.Errorf("failed to disable reaper: %w", err)
		}
	}

	// Use provided password or default
	password := DefaultPassword
	if options.Password != "" {
		password = options.Password
	}

	req := containerRequest(password)
	return testcontainers.GenericContainer(ctx, req)
}

// containerRequest builds the GenericContainerRequest with the resolved password.
// Internal helper; callers should use Run().
func containerRequest(password string) testcontainers.GenericContainerRequest {
	return testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: DefaultImage,
			Env: map[string]string{
				"ACCEPT_EULA": "Y",
				"SA_PASSWORD": password,
			},
			ExposedPorts: []string{defaultPort},
			WaitingFor:   wait.ForListeningPort(defaultPort),
		},
		Started: true,
	}
}

// disableReaperFile writes ~/.testcontainers.properties with ryuk.disabled=true.
func disableReaperFile() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	path := filepath.Join(home, ".testcontainers.properties")
	content := []byte("ryuk.disabled=true\n")

	return os.WriteFile(path, content, 0o644)
}

// ConnectionString returns a fully constructed sqlserver connection string
// using the resolved host, mapped port, and supplied password.
func ConnectionString(ctx context.Context, c testcontainers.Container, password string) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get host: %w", err)
	}

	mapped, err := c.MappedPort(ctx, defaultPort)
	if err != nil {
		return "", fmt.Errorf("failed to get mapped port: %w", err)
	}

	connStr := fmt.Sprintf(
		"sqlserver://%s:%s@%s:%s?database=master&encrypt=disable",
		DefaultUser,
		password,
		host,
		mapped.Port(),
	)

	return connStr, nil
}
