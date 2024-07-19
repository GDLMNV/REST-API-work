package utils

import (
	"context"

	"github.com/GDLMNV/api-mc/pkg/httpErrors"
	"github.com/GDLMNV/api-mc/pkg/logger"
)

func ValidateIsOwner(ctx context.Context, creatorID string, logger logger.Logger) error {
	user, err := GetUserFromCtx(ctx)
	if err != nil {
		return err
	}

	if user.UserID.String() != creatorID {
		logger.Errorf(
			"ValidateIsOwner, userID: %v, creatorID: %v",
			user.UserID.String(),
			creatorID,
		)
		return httpErrors.Forbidden
	}

	return nil
}
