package gomedia

import "io"

type StaticCredentials struct {
	AccessKeyID     string
	SecretAccessKey string
}

type MediaStoreConfig struct {
	BucketName        string
	StaticCredentials *StaticCredentials
	SslEnabled        bool
	Endpoint          string
	Region            *string
	TimeoutSeconds    *uint16
}

type MediaMeta struct {
	Name     string
	MetaData *map[string]string
}

type Media struct {
	MediaMeta
	Reader *io.ReadCloser
}
