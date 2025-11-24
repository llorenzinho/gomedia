package gomedia

import (
	"bytes"
	"context"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioMetaStore struct {
	config MediaStoreConfig
	client *minio.Client
}

func NewMinioMetaStore(config MediaStoreConfig) (*MinioMetaStore, error) {
	opts := minio.Options{
		Secure: config.SslEnabled,
	}
	if config.StaticCredentials != nil {
		opts.Creds = credentials.NewStaticV4(
			config.StaticCredentials.AccessKeyID,
			config.StaticCredentials.SecretAccessKey,
			"",
		)
	}
	client, err := minio.New(config.Endpoint, &opts)
	if err != nil {
		return nil, err
	}

	return &MinioMetaStore{
		config: config,
		client: client,
	}, nil
}

func (m *MinioMetaStore) saveMedia(data *io.Reader, objectPath string, metaData map[string]string) error {
	_, err := m.client.PutObject(
		context.Background(), // TODO: manage context
		m.config.BucketName,
		objectPath,
		*data,
		-1, // TODO: manage size
		minio.PutObjectOptions{
			UserMetadata: metaData,
			ContentType:  "application/octet-astream",
		},
	)
	return err
}

func (m *MinioMetaStore) HealthCheck() error {
	var emptyReader io.Reader = bytes.NewReader([]byte{})
	err := m.saveMedia(&emptyReader, "healthcheck-object", nil)
	return err
}

func (m *MinioMetaStore) SaveMedia(r *io.Reader, meta MediaMeta) (string, error) {
	return "", nil
}

func (m *MinioMetaStore) DeleteMedia(id string) error {
	return nil
}

func (m *MinioMetaStore) GetMediaURL(id string) (*string, error) {
	return nil, nil
}

func (m *MinioMetaStore) GetMediaReader(id string) (*Media, error) {
	return nil, nil
}
