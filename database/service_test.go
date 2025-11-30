package database_test

import (
	"testing"
	"time"

	"github.com/llorenzinho/gomedia/database"
	"github.com/llorenzinho/gomedia/testutils"
	"gorm.io/driver/postgres"
)

func TestDatabaseConnection(t *testing.T) {
	_, dsn, cleanup := testutils.CreateDatabase(t)
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
	_, dsn, cleanup := testutils.CreateDatabase(t)
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
		Check:    true,
	}
	err = dbsvc.CreateMedia(media)
	if err != nil {
		t.Fatalf("failed to create media: %v", err)
	}
	if media.ID == 0 {
		t.Fatalf("expected media ID to be set, got 0")
	}
	retrieved := dbsvc.GetMedia(media.ID)
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
	err = dbsvc.CreateMedia(media2)
	if err != nil {
		t.Fatalf("failed to create media: %v", err)
	}
	err = dbsvc.CreateMedia(media3)
	if err != nil {
		t.Fatalf("failed to create media: %v", err)
	}
	for i, m := range []*database.Media{media2, media3} {
		if m.ID == 0 {
			t.Fatalf("expected media ID to be set for media %d, got 0", i)
		}
	}
}
