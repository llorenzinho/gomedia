package gomedia

import "strconv"

type ErrUnsupportedMediaProvider struct{}

func (e ErrUnsupportedMediaProvider) Error() string {
	return "unsupported media provider"
}

type ErrNilDatabaseService struct{}

func (e ErrNilDatabaseService) Error() string {
	return "database media service cannot be nil"
}

type ErrMediaNotFound struct {
	ID uint
}

func (e ErrMediaNotFound) Error() string {
	return "media not found with ID: " + strconv.FormatUint(uint64(e.ID), 10)
}
