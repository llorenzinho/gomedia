package gomedia

import (
	"io"
)

type mediaSaver interface {
	SaveMedia(r *io.Reader, meta MediaMeta) (string, error) // saves media and return its database id
}

type mediaDeleter interface {
	DeleteMedia(id string) error // deletes media by its database id
}

type mediaURLGetter interface {
	GetMediaURL(id string) (*string, error) // returns a presigned url to access the media
}

type mediaReaderGetter interface {
	GetMediaReader(id string) (*Media, error) // returns a reader to download the media
}

type MediaStorer interface {
	HealthCheck() error
	mediaSaver
	mediaDeleter
	mediaURLGetter
	mediaReaderGetter
}
