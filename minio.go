package gomedia

import (
	"bytes"
	"context"
	"io"
	"log"
	"strconv"

	"github.com/llorenzinho/gomedia/database"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioMetaStore struct {
	config MediaStoreConfig
	client *minio.Client
	db     *database.MediaService
}

func NewMinioMetaStore(config MediaStoreConfig, db *database.MediaService) (*MinioMetaStore, error) {
	if db == nil {
		return nil, ErrNilDatabaseService{}
	}
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
		db:     db,
	}, nil
}

func (m *MinioMetaStore) saveMedia(data *io.Reader, objectPath string, metaData map[string]string, tags map[string]string) error {
	_, err := m.client.PutObject(
		context.Background(), // TODO: manage context
		m.config.BucketName,
		objectPath,
		*data,
		-1, // TODO: manage size
		minio.PutObjectOptions{
			UserMetadata: metaData,
			UserTags:     tags,
			// ContentType:  "application/octet-astream",
		},
	)
	return err
}

func (m *MinioMetaStore) HealthCheck() error {
	var emptyReader io.Reader = bytes.NewReader([]byte{})
	err := m.saveMedia(&emptyReader, "healthcheck-object", nil, nil)
	return err
}

func (m *MinioMetaStore) SaveMedia(r *io.Reader, meta MediaMeta) (*database.Media, error) {
	buf := &bytes.Buffer{}
	size, err := io.Copy(buf, *r)
	if err != nil {
		return nil, err
	}
	mediaEntity := &database.Media{
		Filename: meta.Name,
		Size:     size,
	}
	err = m.db.CreateMedia(mediaEntity)
	if err != nil {
		return nil, err
	}
	err = m.saveMedia(r, strconv.Itoa(int(mediaEntity.ID)), *meta.MetaData, nil)
	if err != nil {
		dms := m.db.DeleteMedia(mediaEntity.ID)
		if len(dms) <= 0 {
			log.Println("Warning: failed to rollback media entity after MinIO upload failure for media ID", mediaEntity.ID)
		}
		return nil, err
	}
	return mediaEntity, nil
}

func (m *MinioMetaStore) DeleteMedia(id uint) error {
	return nil
}

func (m *MinioMetaStore) GetMediaURL(id uint) (*string, error) {
	return nil, nil
}

func (m *MinioMetaStore) GetMediaReader(id uint) (*Media, error) {
	return nil, nil
}
