package gomedia

type ErrUnsupportedMediaProvider struct{}

func (e ErrUnsupportedMediaProvider) Error() string {
	return "unsupported media provider"
}

type ErrNilDatabaseService struct{}

func (e ErrNilDatabaseService) Error() string {
	return "database media service cannot be nil"
}
