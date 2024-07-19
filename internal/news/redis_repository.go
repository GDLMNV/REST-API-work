//go:generate mockgen -source redis_repository.go -destination mock/redis_repository_mock.go -package mock
package news

import (
	"context"

	"github.com/GDLMNV/api-mc/internal/models"
)

type RedisRepository interface {
	GetNewsByIDCtx(ctx context.Context, key string) (*models.NewsBase, error)
	SetNewsCtx(ctx context.Context, key string, seconds int, news *models.NewsBase) error
	DeleteNewsCtx(ctx context.Context, key string) error
}
