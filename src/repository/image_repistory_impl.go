package repository

import (
	"argos/config"
	"context"
	"fmt"
)

const (
	PHHOT        = "photo"
	INVALIDPHOTO = "invalid"
)

func NewImageRepository(database *config.MemoryClient) ImageRepository {
	return &imageRepositoryImpl{
		redis: database,
	}
}

type imageRepositoryImpl struct {
	redis *config.MemoryClient
}

func (i *imageRepositoryImpl) GetInvalidUrl(ctx context.Context, url string) (string, error) {
	key := fmt.Sprintf("%s:%s", INVALIDPHOTO, url)
	return i.redis.Client.Get(ctx, key).Result()
}

func (i *imageRepositoryImpl) CacheInvalidUrl(url string) {
	ctx := context.Background()
	key := fmt.Sprintf("%s:%s", INVALIDPHOTO, url)
	err := i.redis.Client.Set(ctx, key, "invalid", 0).Err()
	fmt.Println(err)
}

func (i *imageRepositoryImpl) HasImage(ctx context.Context, url string) (bool, error) {
	key := fmt.Sprintf("%s:%s", PHHOT, url)
	return i.redis.Client.Get(ctx, key).Bool()
}

func (i *imageRepositoryImpl) GetImage(ctx context.Context, url string) (string, error) {
	key := fmt.Sprintf("%s:%s", PHHOT, url)
	return i.redis.Client.Get(ctx, key).Result()
}

func (i *imageRepositoryImpl) CacheImage(url string, buffer []byte) {
	ctx := context.Background()
	key := fmt.Sprintf("%s:%s", PHHOT, url)
	i.redis.Client.Set(ctx, key, buffer, 0).Err()
}
