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

func (m *MinioMetaStore) saveMedia(data io.Reader, size int64, objectPath string, metaData map[string]string, tags map[string]string) error {
	_, err := m.client.PutObject(
		context.Background(), // TODO: manage context
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
		log.Println("Failed to create media", err)
		return nil, err
	}
	// upload the buffered content; original reader is already consumed
	err = m.saveMedia(bytes.NewReader(buf.Bytes()), size, strconv.Itoa(int(mediaEntity.ID)), meta.MetaData, nil)
	if err != nil {
		log.Println("Failed to upload media in minio: ", err)
		return nil, err
	}
	err = m.db.CheckMedia(mediaEntity.ID)
	if err != nil {
		log.Println("Failed to update media check status: ", err)
		return nil, err
	}

	return mediaEntity, nil
}

func (m *MinioMetaStore) DeleteMedia(id uint) error {
	var media = m.db.GetMedia(id)
	if media == nil {
		return ErrMediaNotFound{ID: id}
	}
	err := m.client.RemoveObject(
		context.Background(), // TODO: manage context
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

func (m *MinioMetaStore) GetMediaReader(id uint) (*Media, error) {
	media := m.db.GetMedia(id)
	if media == nil {
		return nil, ErrMediaNotFound{ID: id}
	}
	object, err := m.client.GetObject(
		context.Background(), // TODO: manage context
		m.config.BucketName,
		strconv.Itoa(int(media.ID)),
		minio.GetObjectOptions{},
	)
	if err != nil {
		return nil, err
	}
	return &Media{
		MediaMeta: MediaMeta{
			Name:     media.Filename,
			MetaData: map[string]string{},
		},
		Reader: object,
	}, nil
}
