package middleware

import (
	"github.com/GDLMNV/api-mc/config"
	"github.com/GDLMNV/api-mc/internal/auth"
	"github.com/GDLMNV/api-mc/internal/session"
	"github.com/GDLMNV/api-mc/pkg/logger"
)

type MiddlewareManager struct {
	sessUC  session.UCSession
	authUC  auth.UseCase
	cfg     *config.Config
	origins []string
	logger  logger.Logger
}

func NewMiddlewareManager(sessUC session.UCSession, authUC auth.UseCase, cfg *config.Config, origins []string, logger logger.Logger) *MiddlewareManager {
	return &MiddlewareManager{sessUC: sessUC, authUC: authUC, cfg: cfg, origins: origins, logger: logger}
}
