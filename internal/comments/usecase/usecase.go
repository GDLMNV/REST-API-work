package usecase

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"

	"github.com/GDLMNV/api-mc/config"
	"github.com/GDLMNV/api-mc/internal/comments"
	"github.com/GDLMNV/api-mc/internal/models"
	"github.com/GDLMNV/api-mc/pkg/httpErrors"
	"github.com/GDLMNV/api-mc/pkg/logger"
	"github.com/GDLMNV/api-mc/pkg/utils"
)

type commentsUC struct {
	cfg      *config.Config
	commRepo comments.Repository
	logger   logger.Logger
}

func NewCommentsUseCase(cfg *config.Config, commRepo comments.Repository, logger logger.Logger) comments.UseCase {
	return &commentsUC{cfg: cfg, commRepo: commRepo, logger: logger}
}

func (u *commentsUC) Create(ctx context.Context, comment *models.Comment) (*models.Comment, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "commentsUC.Create")
	defer span.Finish()
	return u.commRepo.Create(ctx, comment)
}

func (u *commentsUC) Update(ctx context.Context, comment *models.Comment) (*models.Comment, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "commentsUC.Update")
	defer span.Finish()

	comm, err := u.commRepo.GetByID(ctx, comment.CommentID)
	if err != nil {
		return nil, err
	}

	if err = utils.ValidateIsOwner(ctx, comm.AuthorID.String(), u.logger); err != nil {
		return nil, httpErrors.NewRestError(http.StatusForbidden, "Forbidden", errors.Wrap(err, "commentsUC.Update.ValidateIsOwner"))
	}

	updatedComment, err := u.commRepo.Update(ctx, comment)
	if err != nil {
		return nil, err
	}

	return updatedComment, nil
}

func (u *commentsUC) Delete(ctx context.Context, commentID uuid.UUID) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "commentsUC.Delete")
	defer span.Finish()

	comm, err := u.commRepo.GetByID(ctx, commentID)
	if err != nil {
		return err
	}

	if err = utils.ValidateIsOwner(ctx, comm.AuthorID.String(), u.logger); err != nil {
		return httpErrors.NewRestError(http.StatusForbidden, "Forbidden", errors.Wrap(err, "commentsUC.Delete.ValidateIsOwner"))
	}

	if err = u.commRepo.Delete(ctx, commentID); err != nil {
		return err
	}

	return nil
}

func (u *commentsUC) GetByID(ctx context.Context, commentID uuid.UUID) (*models.CommentBase, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "commentsUC.GetByID")
	defer span.Finish()

	return u.commRepo.GetByID(ctx, commentID)
}

func (u *commentsUC) GetAllByNewsID(ctx context.Context, newsID uuid.UUID, query *utils.PaginationQuery) (*models.CommentsList, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "commentsUC.GetAllByNewsID")
	defer span.Finish()

	return u.commRepo.GetAllByNewsID(ctx, newsID, query)
}
