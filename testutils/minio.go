package testutils

import (
	"context"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func CreateMinioStore(t *testing.T) (*testcontainers.DockerContainer, func()) {
	ctx := context.Background()
	minio, err := testcontainers.Run(
		ctx,
		"minio/minio:latest",
		testcontainers.WithEnv(map[string]string{
			"MINIO_ROOT_USER":     "minioadmin",
			"MINIO_ROOT_PASSWORD": "minioadmin",
		}),
		testcontainers.WithCmd("server", "/data", "--console-address", ":9001"),
		testcontainers.WithExposedPorts("9000/tcp", "9001/tcp"),
		testcontainers.WithWaitStrategy(wait.ForListeningPort("9001/tcp")),
	)
	if err != nil {
		t.Fatalf("failed to start minio container: %v", err)
	}

	bucketName := "test-bucket"
	aliasName := "local"
	minioURL := "http://127.0.0.1:9000"
	accessKey := "minioadmin"
	secretKey := "minioadmin"

	aliasCmd := []string{"mc", "alias", "set", aliasName, minioURL, accessKey, secretKey}
	if _, _, err := minio.Exec(ctx, aliasCmd); err != nil {
		t.Fatalf("failed to set MinIO alias: %v", err)
	}

	createBucketCmd := []string{"mc", "mb", aliasName + "/" + bucketName}
	if _, _, err := minio.Exec(ctx, createBucketCmd); err != nil {
		t.Fatalf("failed to create bucket: %v", err)
	}

	cleanup := func() {
		_ = minio.Terminate(ctx)
	}
	return minio, cleanup
}
