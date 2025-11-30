package gomedia

import (
	"io"
)

type mediaSaver interface {
	SaveMedia(r *io.Reader, meta MediaMeta) (*MediaEntity, error) // saves media and return its database id
}

type mediaDeleter interface {
	DeleteMedia(id uint) error // deletes media by its database id
}

type mediaGetter interface {
	GetMedia(id uint) (*Media, error) // returns a reader to download the media
}

type MediaStorer interface {
	HealthCheck() error
	mediaSaver
	mediaDeleter
	mediaGetter
}
