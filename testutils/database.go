package testutils

import (
	"context"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func CreateDatabase(t *testing.T) (*testcontainers.DockerContainer, string, func()) {
	ctx := context.Background()
	db, err := testcontainers.Run(
		ctx,
		"postgres:latest",
		testcontainers.WithEnv(map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "testdb",
		}),
		testcontainers.WithExposedPorts("5432/tcp"),
		testcontainers.WithWaitStrategy(wait.ForListeningPort("5432/tcp")),
	)
	if err != nil {
		t.Fatalf("failed to start database container: %v", err)
	}

	host, err := db.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get database host: %v", err)
	}
	port, err := db.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatalf("failed to get database port: %v", err)
	}

	dsn := "host=" + host + " port=" + port.Port() + " user=testuser password=testpass dbname=testdb sslmode=disable"

	return db, dsn, func() {
		_ = db.Terminate(ctx)
	}
}
