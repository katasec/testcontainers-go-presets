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

// -----------------------------------------------------------------------------
// Options
// -----------------------------------------------------------------------------

// RunOptions holds optional flags and overrides for Run.
type RunOptions struct {
	DisableReaper bool
	Password      string
}

// RunOption configures RunOptions.
type RunOption func(*RunOptions)

func WithReaperDisabled() RunOption {
	return func(o *RunOptions) {
		o.DisableReaper = true
	}
}

func WithPassword(pw string) RunOption {
	return func(o *RunOptions) {
		o.Password = pw
	}
}

// -----------------------------------------------------------------------------
// Run
// -----------------------------------------------------------------------------

// Run starts an MSSQL container with optional RunOptions.
func Run(ctx context.Context, opts ...RunOption) (testcontainers.Container, error) {
	// Collect options
	options := RunOptions{}
	for _, optionFunc := range opts {
		optionFunc(&options)
	}

	// Disable Ryuk global reaper if requested
	if options.DisableReaper {
		if err := disableReaperFile(); err != nil {
			return nil, fmt.Errorf("failed to disable reaper: %w", err)
		}
	}

	password := DefaultPassword
	if options.Password != "" {
		password = options.Password
	}

	// Build container request with mounted config file
	req := containerRequest(password)

	return testcontainers.GenericContainer(ctx, req)
}

// -----------------------------------------------------------------------------
// Internal: container request builder
// -----------------------------------------------------------------------------

// containerRequest builds the GenericContainerRequest with SQL Agent enabled.
// It *always* generates and mounts mssql.conf before container creation.
func containerRequest(password string) testcontainers.GenericContainerRequest {
	// Create temp config file (enables SQL Agent)
	confPath, err := createMSSQLConf()
	if err != nil {
		// Fail-fast: presets must not silently misconfigure SQL Server
		panic(fmt.Errorf("failed to create mssql.conf: %w", err))
	}

	// File mount instruction
	files := []testcontainers.ContainerFile{
		{
			HostFilePath:      confPath,
			ContainerFilePath: "/var/opt/mssql/mssql.conf",
			FileMode:          0o644,
		},
	}

	return testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: DefaultImage,
			Env: map[string]string{
				"ACCEPT_EULA": "Y",
				"SA_PASSWORD": password,
			},
			ExposedPorts: []string{defaultPort},
			WaitingFor:   wait.ForListeningPort(defaultPort),
			Files:        files,
		},
		Started: true,
	}
}

// -----------------------------------------------------------------------------
// Internal helpers
// -----------------------------------------------------------------------------

// createMSSQLConf enables SQL Agent so CDC works inside the container.
func createMSSQLConf() (string, error) {
	dir, err := os.MkdirTemp("", "mssql-conf-*")
	if err != nil {
		return "", err
	}

	path := filepath.Join(dir, "mssql.conf")

	content := []byte("[sqlagent]\nenabled = true\n")

	if err := os.WriteFile(path, content, 0o644); err != nil {
		return "", err
	}

	return path, nil
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

// -----------------------------------------------------------------------------
// Public: Connection string
// -----------------------------------------------------------------------------

// ConnectionString builds a sqlserver:// connection string for the given DB name.
// If database = "", "master" is used.
func ConnectionString(ctx context.Context, c testcontainers.Container, password string, database string) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get host: %w", err)
	}

	mapped, err := c.MappedPort(ctx, defaultPort)
	if err != nil {
		return "", fmt.Errorf("failed to get mapped port: %w", err)
	}

	if database == "" {
		database = "master"
	}

	connStr := fmt.Sprintf(
		"sqlserver://%s:%s@%s:%s?database=%s&encrypt=disable",
		DefaultUser,
		password,
		host,
		mapped.Port(),
		database,
	)

	return connStr, nil
}
