package gomedia

type Option func(*MediaStoreConfig)

func WithStaticCredentials(accessKeyID, secretAccessKey string) Option {
	return func(config *MediaStoreConfig) {
		config.StaticCredentials = &StaticCredentials{
			AccessKeyID:     accessKeyID,
			SecretAccessKey: secretAccessKey,
		}
	}
}

func WithSSLEnabled(config *MediaStoreConfig) {
	config.SslEnabled = true
}

func WithEndpoint(endpoint string) Option {
	return func(config *MediaStoreConfig) {
		config.Endpoint = endpoint
	}
}

func WithRegion(region string) Option {
	return func(config *MediaStoreConfig) {
		config.Region = &region
	}
}

func WithTimeoutSeconds(timeout uint16) Option {
	return func(config *MediaStoreConfig) {
		config.TimeoutSeconds = timeout
	}
}

func WithContinuousHealthCheck(intervalSeconds uint16) Option {
	return func(config *MediaStoreConfig) {
		config.ContinuousHealthCheckSeconds = &intervalSeconds
	}
}

func NewMediaStore(
	provider mediaProvider,
	bucket string,
	db *mediaService,
	opts ...Option,
) (MediaStorer, error) {
	config := &MediaStoreConfig{
		BucketName: bucket,
		SslEnabled: false,
	}
	for _, opt := range opts {
		opt(config)
	}
	switch provider {
	case MediaProviderMinio:
		return NewMinioMetaStore(*config, db)
	default:
		return nil, ErrUnsupportedMediaProvider{}
	}

}
