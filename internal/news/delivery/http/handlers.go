package http

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/opentracing/opentracing-go"

	"github.com/GDLMNV/api-mc/config"
	"github.com/GDLMNV/api-mc/internal/models"
	"github.com/GDLMNV/api-mc/internal/news"
	"github.com/GDLMNV/api-mc/pkg/httpErrors"
	"github.com/GDLMNV/api-mc/pkg/logger"
	"github.com/GDLMNV/api-mc/pkg/utils"
)

type newsHandlers struct {
	cfg    *config.Config
	newsUC news.UseCase
	logger logger.Logger
}

func NewNewsHandlers(cfg *config.Config, newsUC news.UseCase, logger logger.Logger) news.Handlers {
	return &newsHandlers{cfg: cfg, newsUC: newsUC, logger: logger}
}

func (h newsHandlers) Create() echo.HandlerFunc {
	return func(c echo.Context) error {
		span, ctx := opentracing.StartSpanFromContext(utils.GetRequestCtx(c), "newsHandlers.Create")
		defer span.Finish()

		n := &models.News{}
		if err := c.Bind(n); err != nil {
			utils.LogResponseError(c, h.logger, err)
			return c.JSON(httpErrors.ErrorResponse(err))
		}

		createdNews, err := h.newsUC.Create(ctx, n)
		if err != nil {
			utils.LogResponseError(c, h.logger, err)
			return c.JSON(httpErrors.ErrorResponse(err))
		}

		return c.JSON(http.StatusCreated, createdNews)
	}
}

func (h newsHandlers) Update() echo.HandlerFunc {
	return func(c echo.Context) error {
		span, ctx := opentracing.StartSpanFromContext(utils.GetRequestCtx(c), "newsHandlers.Update")
		defer span.Finish()

		newsUUID, err := uuid.Parse(c.Param("news_id"))
		if err != nil {
			utils.LogResponseError(c, h.logger, err)
			return c.JSON(httpErrors.ErrorResponse(err))
		}

		n := &models.News{}
		if err = c.Bind(n); err != nil {
			utils.LogResponseError(c, h.logger, err)
			return c.JSON(httpErrors.ErrorResponse(err))
		}
		n.NewsID = newsUUID

		updatedNews, err := h.newsUC.Update(ctx, n)
		if err != nil {
			utils.LogResponseError(c, h.logger, err)
			return c.JSON(httpErrors.ErrorResponse(err))
		}

		return c.JSON(http.StatusOK, updatedNews)
	}
}

func (h newsHandlers) GetByID() echo.HandlerFunc {
	return func(c echo.Context) error {
		span, ctx := opentracing.StartSpanFromContext(utils.GetRequestCtx(c), "newsHandlers.GetByID")
		defer span.Finish()

		newsUUID, err := uuid.Parse(c.Param("news_id"))
		if err != nil {
			utils.LogResponseError(c, h.logger, err)
			return c.JSON(httpErrors.ErrorResponse(err))
		}

		newsByID, err := h.newsUC.GetNewsByID(ctx, newsUUID)
		if err != nil {
			utils.LogResponseError(c, h.logger, err)
			return c.JSON(httpErrors.ErrorResponse(err))
		}

		return c.JSON(http.StatusOK, newsByID)
	}
}

func (h newsHandlers) Delete() echo.HandlerFunc {
	return func(c echo.Context) error {
		span, ctx := opentracing.StartSpanFromContext(utils.GetRequestCtx(c), "newsHandlers.Delete")
		defer span.Finish()

		newsUUID, err := uuid.Parse(c.Param("news_id"))
		if err != nil {
			utils.LogResponseError(c, h.logger, err)
			return c.JSON(httpErrors.ErrorResponse(err))
		}

		if err = h.newsUC.Delete(ctx, newsUUID); err != nil {
			utils.LogResponseError(c, h.logger, err)
			return c.JSON(httpErrors.ErrorResponse(err))
		}

		return c.NoContent(http.StatusOK)
	}
}

func (h newsHandlers) GetNews() echo.HandlerFunc {
	return func(c echo.Context) error {
		span, ctx := opentracing.StartSpanFromContext(utils.GetRequestCtx(c), "newsHandlers.GetNews")
		defer span.Finish()

		pq, err := utils.GetPaginationFromCtx(c)
		if err != nil {
			utils.LogResponseError(c, h.logger, err)
			return c.JSON(httpErrors.ErrorResponse(err))
		}

		newsList, err := h.newsUC.GetNews(ctx, pq)
		if err != nil {
			utils.LogResponseError(c, h.logger, err)
			return c.JSON(httpErrors.ErrorResponse(err))
		}

		return c.JSON(http.StatusOK, newsList)
	}
}

func (h newsHandlers) SearchByTitle() echo.HandlerFunc {
	return func(c echo.Context) error {
		span, ctx := opentracing.StartSpanFromContext(utils.GetRequestCtx(c), "newsHandlers.SearchByTitle")
		defer span.Finish()

		pq, err := utils.GetPaginationFromCtx(c)
		if err != nil {
			utils.LogResponseError(c, h.logger, err)
			return c.JSON(httpErrors.ErrorResponse(err))
		}

		newsList, err := h.newsUC.SearchByTitle(ctx, c.QueryParam("title"), pq)

		if err != nil {
			utils.LogResponseError(c, h.logger, err)
			return c.JSON(httpErrors.ErrorResponse(err))
		}

		return c.JSON(http.StatusOK, newsList)
	}
}
