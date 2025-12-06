package mssql

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	DefaultImage    = "mcr.microsoft.com/mssql/server:2022-latest"
	DefaultUser     = "sa"
	DefaultPassword = "Your_password123!" // for tests only
	defaultPort     = "1433/tcp"
)

// ContainerRequest returns a ready-to-use GenericContainerRequest
// for running SQL Server in tests.
func ContainerRequest() testcontainers.GenericContainerRequest {
	return testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: DefaultImage,
			Env: map[string]string{
				"ACCEPT_EULA": "Y",
				"SA_PASSWORD": DefaultPassword,
			},
			ExposedPorts: []string{defaultPort},
			WaitingFor:   wait.ForListeningPort(defaultPort),
		},
		Started: true,
	}
}

// Run starts an MSSQL container with default settings.
func Run(ctx context.Context) (testcontainers.Container, error) {
	req := ContainerRequest()
	return testcontainers.GenericContainer(ctx, req)
}
