package database_test

import (
	"context"
	"testing"
	"time"

	"github.com/llorenzinho/gomedia/database"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
)

func createDatabase(t *testing.T) (*testcontainers.DockerContainer, string, func()) {
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

func TestDatabaseConnection(t *testing.T) {
	_, dsn, cleanup := createDatabase(t)
	defer cleanup()

	dbsvc := database.NewMediaService(
		postgres.Open(dsn),
		database.WithPoolMaxLifetime(10*time.Minute),
		database.WithPoolMaxIdleConns(10),
		database.WithPoolMaxOpenConns(100),
	)
	err := dbsvc.AutoMigrate()
	if err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}
	err = dbsvc.Ping()
	if err != nil {
		t.Fatalf("failed to ping database: %v", err)
	}
}

func TestMediaCreation(t *testing.T) {
	_, dsn, cleanup := createDatabase(t)
	defer cleanup()

	dbsvc := database.NewMediaService(
		postgres.Open(dsn),
	)
	err := dbsvc.AutoMigrate()
	if err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}
	media := &database.Media{
		Filename: "testfile.jpg",
		Size:     2048,
	}
	result := dbsvc.CreateMedia(media)
	if len(result) != 1 {
		t.Fatalf("expected 1 media to be created, got %d", len(result))
	}
	if result[0].ID == 0 {
		t.Fatalf("expected media ID to be set, got 0")
	}
	retrieved := dbsvc.GetMedia(result[0].ID)
	if retrieved == nil {
		t.Fatalf("failed to retrieve created media")
	}
	if retrieved.Filename != "testfile.jpg" {
		t.Fatalf("expected filename 'testfile.jpg', got '%s'", retrieved.Filename)
	}
	if retrieved.Size != 2048 {
		t.Fatalf("expected size 2048, got %d", retrieved.Size)
	}

	// Create multiple media
	media2 := &database.Media{
		Filename: "file2.png",
		Size:     4096,
	}
	media3 := &database.Media{
		Filename: "file3.mp4",
		Size:     8192,
	}
	results := dbsvc.CreateMedia(media2, media3)
	if len(results) != 2 {
		t.Fatalf("expected 2 media to be created, got %d", len(results))
	}
	for i, media := range results {
		if media.ID == 0 {
			t.Fatalf("expected media ID to be set for media %d, got 0", i)
		}
	}
}
