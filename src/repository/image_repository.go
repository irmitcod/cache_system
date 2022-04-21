package repository

import "context"

type ImageRepository interface {
	CacheImage(url string, buffer []byte)
	CacheInvalidUrl(url string)
	GetInvalidUrl(ctx context.Context, url string) (string, error)
	GetImage(ctx context.Context, url string) (string, error)
	HasImage(ctx context.Context, url string) (bool, error)
}
