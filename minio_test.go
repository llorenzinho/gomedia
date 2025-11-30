package gomedia

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/llorenzinho/gomedia/testutils"
	"gorm.io/driver/postgres"
)

func TestMinioHealth(t *testing.T) {
	ctx := context.Background()
	minio, cleanup := testutils.CreateMinioStore(t)
	defer cleanup()
	_, dsn, cleanupDb := testutils.CreateDatabase(t)
	defer cleanupDb()
	dbsvc := newMediaService(
		postgres.Open(dsn),
		withPoolMaxLifetime(10*time.Minute),
		withPoolMaxIdleConns(10),
		withPoolMaxOpenConns(100),
	)

	endpoint, err := minio.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get minio host: %v", err)
	}
	port, err := minio.MappedPort(ctx, "9000")
	if err != nil {
		t.Fatalf("failed to get minio port: %v", err)
	}

	m, err := NewMediaStore(
		MediaProviderMinio,
		"test-bucket",
		dbsvc,
		WithEndpoint(endpoint+":"+port.Port()),
		WithStaticCredentials("minioadmin", "minioadmin"),
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
	dbsvc := newMediaService(
		postgres.Open(dsn),
		withPoolMaxLifetime(10*time.Minute),
		withPoolMaxIdleConns(10),
		withPoolMaxOpenConns(100),
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

	m, err := NewMediaStore(
		MediaProviderMinio,
		"test-bucket",
		dbsvc,
		WithEndpoint(endpoint+":"+port.Port()),
		WithStaticCredentials("minioadmin", "minioadmin"),
	)
	if err != nil {
		t.Fatalf("failed to create minio media store: %v", err)
	}

	content := "Hello, MinIO!"
	media := Media{
		MediaMeta: MediaMeta{
			Name: "hello.txt",
		},
		Reader: strings.NewReader(content),
	}

	cm, err := m.SaveMedia(&media.Reader, media.MediaMeta)
	if err != nil {
		t.Fatalf("failed to save media: %v", err)
	}

	rm, err := m.GetMedia(cm.ID)
	if err != nil {
		t.Fatalf("failed to get media reader: %v", err)
	}

	buf := new(strings.Builder)
	_, err = io.Copy(buf, rm.Reader)
	if err != nil {
		t.Fatalf("failed to read media content: %v", err)
	}

	if buf.String() != content {
		t.Fatalf("media content mismatch: expected %q, got %q", content, buf.String())
	}
}

func TestMinioDelete(t *testing.T) {
	ctx := context.Background()
	minio, cleanup := testutils.CreateMinioStore(t)
	defer cleanup()

	_, dsn, cleanupDb := testutils.CreateDatabase(t)
	defer cleanupDb()
	dbsvc := newMediaService(
		postgres.Open(dsn),
		withPoolMaxLifetime(10*time.Minute),
		withPoolMaxIdleConns(10),
		withPoolMaxOpenConns(100),
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

	m, err := NewMediaStore(
		MediaProviderMinio,
		"test-bucket",
		dbsvc,
		WithEndpoint(endpoint+":"+port.Port()),
		WithStaticCredentials("minioadmin", "minioadmin"),
	)
	if err != nil {
		t.Fatalf("failed to create minio media store: %v", err)
	}

	content := "Hello, MinIO!"
	media := Media{
		MediaMeta: MediaMeta{
			Name: "hello.txt",
		},
		Reader: strings.NewReader(content),
	}

	cm, err := m.SaveMedia(&media.Reader, media.MediaMeta)
	if err != nil {
		t.Fatalf("failed to save media: %v", err)
	}

	err = m.DeleteMedia(cm.ID)
	if err != nil {
		t.Fatalf("failed to delete media: %v", err)
	}

	_, err = m.GetMedia(cm.ID)
	if err == nil {
		t.Fatalf("expected error when getting deleted media, got none")
	}
}
