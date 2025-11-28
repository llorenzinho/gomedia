package gomedia_test

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/llorenzinho/gomedia"
	"github.com/llorenzinho/gomedia/database"
	"github.com/llorenzinho/gomedia/testutils"
	"gorm.io/driver/postgres"
)

func TestMinioHealth(t *testing.T) {
	ctx := context.Background()
	minio, cleanup := testutils.CreateMinioStore(t)
	defer cleanup()
	_, dsn, cleanupDb := testutils.CreateDatabase(t)
	defer cleanupDb()
	dbsvc := database.NewMediaService(
		postgres.Open(dsn),
		database.WithPoolMaxLifetime(10*time.Minute),
		database.WithPoolMaxIdleConns(10),
		database.WithPoolMaxOpenConns(100),
	)

	endpoint, err := minio.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get minio host: %v", err)
	}
	port, err := minio.MappedPort(ctx, "9000")
	if err != nil {
		t.Fatalf("failed to get minio port: %v", err)
	}

	m, err := gomedia.NewMediaStore(
		gomedia.MediaProviderMinio,
		"test-bucket",
		dbsvc,
		gomedia.WithEndpoint(endpoint+":"+port.Port()),
		gomedia.WithStaticCredentials("minioadmin", "minioadmin"),
	)
	if err != nil {
		t.Fatalf("failed to create minio media store: %v", err)
	}
	err = m.HealthCheck()
	if err != nil {
		t.Fatalf("minio health check failed: %v", err)
	}
}

func TestMinioUpload(t *testing.T) {
	ctx := context.Background()
	minio, cleanup := testutils.CreateMinioStore(t)
	defer cleanup()

	_, dsn, cleanupDb := testutils.CreateDatabase(t)
	defer cleanupDb()
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

	endpoint, err := minio.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get minio host: %v", err)
	}
	port, err := minio.MappedPort(ctx, "9000")
	if err != nil {
		t.Fatalf("failed to get minio port: %v", err)
	}

	m, err := gomedia.NewMediaStore(
		gomedia.MediaProviderMinio,
		"test-bucket",
		dbsvc,
		gomedia.WithEndpoint(endpoint+":"+port.Port()),
		gomedia.WithStaticCredentials("minioadmin", "minioadmin"),
	)
	if err != nil {
		t.Fatalf("failed to create minio media store: %v", err)
	}

	content := "Hello, MinIO!"
	media := gomedia.Media{
		MediaMeta: gomedia.MediaMeta{
			Name: "hello.txt",
		},
		Reader: strings.NewReader(content),
	}

	cm, err := m.SaveMedia(&media.Reader, media.MediaMeta)
	if err != nil {
		t.Fatalf("failed to save media: %v", err)
	}

	retrievedMedia, err := m.GetMediaReader(cm.ID)
	if err != nil {
		t.Fatalf("failed to get media reader: %v", err)
	}

	buf := new(strings.Builder)
	_, err = io.Copy(buf, retrievedMedia.Reader)
	if err != nil {
		t.Fatalf("failed to read media content: %v", err)
	}

	if buf.String() != content {
		t.Fatalf("media content mismatch: expected %q, got %q", content, buf.String())
	}
}
