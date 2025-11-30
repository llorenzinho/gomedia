package gomedia

import (
	"bytes"
	"context"
	"io"
	"strconv"
	"time"

	"github.com/llorenzinho/gomedia/database"
	"github.com/llorenzinho/gomedia/internal"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/phuslu/log"
)

type MinioMetaStore struct {
	config MediaStoreConfig
	client *minio.Client
	db     *database.MediaService
	l      *log.Logger
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

	if config.TimeoutSeconds == 0 {
		config.TimeoutSeconds = 30
	}

	store := &MinioMetaStore{
		config: config,
		client: client,
		db:     db,
		l:      internal.GetLogger(),
	}

	// If continuous health check is enabled, start a goroutine to perform it
	if config.ContinuousHealthCheckSeconds != nil {
		interval := *config.ContinuousHealthCheckSeconds
		go func() {
			for {
				err := store.HealthCheck()
				if err != nil {
					store.l.Error().Err(err).Msg("Minio health check failed")
				}
				time.Sleep(time.Duration(interval) * time.Second)
			}
		}()
	}

	return store, nil
}

func (m *MinioMetaStore) saveMedia(data io.Reader, size int64, objectPath string, metaData map[string]string, tags map[string]string) error {
	ctx, df := context.WithTimeout(context.Background(), time.Duration(m.config.TimeoutSeconds)*time.Second)
	defer df()
	_, err := m.client.PutObject(
		ctx,
		m.config.BucketName,
		objectPath,
		data,
		size,
		minio.PutObjectOptions{
			UserMetadata: metaData,
			UserTags:     tags,
		},
	)
	return err
}

func (m *MinioMetaStore) HealthCheck() error {
	var emptyReader io.Reader = bytes.NewReader([]byte{})
	err := m.saveMedia(emptyReader, 0, "healthcheck-object", nil, nil)
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
		Check:    false,
	}
	err = m.db.CreateMedia(mediaEntity)
	if err != nil {
		m.l.Error().Err(err).Msg("Failed to create media")
		return nil, err
	}
	// upload the buffered content; original reader is already consumed
	err = m.saveMedia(bytes.NewReader(buf.Bytes()), size, strconv.Itoa(int(mediaEntity.ID)), meta.MetaData, nil)
	if err != nil {
		m.l.Error().Err(err).Msg("Failed to upload media in minio: ")
		return nil, err
	}
	err = m.db.CheckMedia(mediaEntity.ID)
	if err != nil {
		m.l.Error().Err(err).Msg("Failed to update media check status: ")
		return nil, err
	}

	return mediaEntity, nil
}

func (m *MinioMetaStore) DeleteMedia(id uint) error {
	var media = m.db.GetMedia(id)
	if media == nil {
		return ErrMediaNotFound{ID: id}
	}
	ctx, df := context.WithTimeout(context.Background(), time.Duration(m.config.TimeoutSeconds)*time.Second)
	defer df()
	err := m.client.RemoveObject(
		ctx,
		m.config.BucketName,
		strconv.Itoa(int(media.ID)),
		minio.RemoveObjectOptions{},
	)
	if err != nil {
		return err
	}
	medias := m.db.DeleteMedias(id)
	if len(medias) == 0 {
		return ErrMediaNotFound{ID: id}
	}
	return nil
}

func (m *MinioMetaStore) GetMedia(id uint) (*Media, error) {
	media := m.db.GetMedia(id)
	if media == nil {
		return nil, ErrMediaNotFound{ID: id}
	}
	ctx, df := context.WithTimeout(context.Background(), time.Duration(m.config.TimeoutSeconds)*time.Second)
	defer df()
	object, err := m.client.GetObject(
		ctx,
		m.config.BucketName,
		strconv.Itoa(int(media.ID)),
		minio.GetObjectOptions{},
	)
	if err != nil {
		return nil, err
	}

	buffer := bytes.Buffer{}
	_, err = io.Copy(&buffer, object)
	if err != nil {
		return nil, err
	}
	return &Media{
		MediaMeta: MediaMeta{
			Name:     media.Filename,
			MetaData: map[string]string{},
		},
		Reader: bytes.NewReader(buffer.Bytes()),
	}, nil
}
