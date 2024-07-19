package http

import (
	"bytes"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"

	"github.com/GDLMNV/api-mc/config"
	"github.com/GDLMNV/api-mc/internal/auth"
	"github.com/GDLMNV/api-mc/internal/models"
	"github.com/GDLMNV/api-mc/internal/session"
	"github.com/GDLMNV/api-mc/pkg/csrf"
	"github.com/GDLMNV/api-mc/pkg/httpErrors"
	"github.com/GDLMNV/api-mc/pkg/logger"
	"github.com/GDLMNV/api-mc/pkg/utils"
)

type authHandlers struct {
	cfg    *config.Config
	authUC auth.UseCase
	sessUC session.UCSession
	logger logger.Logger
}

func NewAuthHandlers(cfg *config.Config, authUC auth.UseCase, sessUC session.UCSession, log logger.Logger) auth.Handlers {
	return &authHandlers{cfg: cfg, authUC: authUC, sessUC: sessUC, logger: log}
}

func (h *authHandlers) Register() echo.HandlerFunc {
	return func(c echo.Context) error {
		span, ctx := opentracing.StartSpanFromContext(utils.GetRequestCtx(c), "auth.Register")
		defer span.Finish()

		user := &models.User{}
		if err := utils.ReadRequest(c, user); err != nil {
			utils.LogResponseError(c, h.logger, err)
			return c.JSON(httpErrors.ErrorResponse(err))
		}

		createdUser, err := h.authUC.Register(ctx, user)
		if err != nil {
			utils.LogResponseError(c, h.logger, err)
			return c.JSON(httpErrors.ErrorResponse(err))
		}

		sess, err := h.sessUC.CreateSession(ctx, &models.Session{
			UserID: createdUser.User.UserID,
		}, h.cfg.Session.Expire)
		if err != nil {
			utils.LogResponseError(c, h.logger, err)
			return c.JSON(httpErrors.ErrorResponse(err))
		}

		c.SetCookie(utils.CreateSessionCookie(h.cfg, sess))

		return c.JSON(http.StatusCreated, createdUser)
	}
}

func (h *authHandlers) Login() echo.HandlerFunc {
	type Login struct {
		Email    string `json:"email" db:"email" validate:"omitempty,lte=60,email"`
		Password string `json:"password,omitempty" db:"password" validate:"required,gte=6"`
	}
	return func(c echo.Context) error {
		span, ctx := opentracing.StartSpanFromContext(utils.GetRequestCtx(c), "auth.Login")
		defer span.Finish()

		login := &Login{}
		if err := utils.ReadRequest(c, login); err != nil {
			utils.LogResponseError(c, h.logger, err)
			return c.JSON(httpErrors.ErrorResponse(err))
		}

		userWithToken, err := h.authUC.Login(ctx, &models.User{
			Email:    login.Email,
			Password: login.Password,
		})
		if err != nil {
			utils.LogResponseError(c, h.logger, err)
			return c.JSON(httpErrors.ErrorResponse(err))
		}

		sess, err := h.sessUC.CreateSession(ctx, &models.Session{
			UserID: userWithToken.User.UserID,
		}, h.cfg.Session.Expire)
		if err != nil {
			utils.LogResponseError(c, h.logger, err)
			return c.JSON(httpErrors.ErrorResponse(err))
		}

		c.SetCookie(utils.CreateSessionCookie(h.cfg, sess))

		return c.JSON(http.StatusOK, userWithToken)
	}
}

func (h *authHandlers) Logout() echo.HandlerFunc {
	return func(c echo.Context) error {
		span, ctx := opentracing.StartSpanFromContext(utils.GetRequestCtx(c), "authHandlers.Logout")
		defer span.Finish()

		cookie, err := c.Cookie("session-id")
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				utils.LogResponseError(c, h.logger, err)
				return c.JSON(http.StatusUnauthorized, httpErrors.NewUnauthorizedError(err))
			}
			utils.LogResponseError(c, h.logger, err)
			return c.JSON(http.StatusInternalServerError, httpErrors.NewInternalServerError(err))
		}

		if err := h.sessUC.DeleteByID(ctx, cookie.Value); err != nil {
			utils.LogResponseError(c, h.logger, err)
			return c.JSON(httpErrors.ErrorResponse(err))
		}

		utils.DeleteSessionCookie(c, h.cfg.Session.Name)

		return c.NoContent(http.StatusOK)
	}
}

func (h *authHandlers) Update() echo.HandlerFunc {
	return func(c echo.Context) error {
		span, ctx := opentracing.StartSpanFromContext(utils.GetRequestCtx(c), "authHandlers.Update")
		defer span.Finish()

		uID, err := uuid.Parse(c.Param("user_id"))
		if err != nil {
			utils.LogResponseError(c, h.logger, err)
			return c.JSON(httpErrors.ErrorResponse(err))
		}

		user := &models.User{}
		user.UserID = uID

		if err = utils.ReadRequest(c, user); err != nil {
			utils.LogResponseError(c, h.logger, err)
			return c.JSON(httpErrors.ErrorResponse(err))
		}

		updatedUser, err := h.authUC.Update(ctx, user)
		if err != nil {
			utils.LogResponseError(c, h.logger, err)
			return c.JSON(httpErrors.ErrorResponse(err))
		}

		return c.JSON(http.StatusOK, updatedUser)
	}
}

func (h *authHandlers) GetUserByID() echo.HandlerFunc {
	return func(c echo.Context) error {
		span, ctx := opentracing.StartSpanFromContext(utils.GetRequestCtx(c), "authHandlers.GetUserByID")
		defer span.Finish()

		uID, err := uuid.Parse(c.Param("user_id"))
		if err != nil {
			utils.LogResponseError(c, h.logger, err)
			return c.JSON(httpErrors.ErrorResponse(err))
		}

		user, err := h.authUC.GetByID(ctx, uID)
		if err != nil {
			utils.LogResponseError(c, h.logger, err)
			return c.JSON(httpErrors.ErrorResponse(err))
		}

		return c.JSON(http.StatusOK, user)
	}
}

func (h *authHandlers) Delete() echo.HandlerFunc {
	return func(c echo.Context) error {
		span, ctx := opentracing.StartSpanFromContext(utils.GetRequestCtx(c), "authHandlers.Delete")
		defer span.Finish()

		uID, err := uuid.Parse(c.Param("user_id"))
		if err != nil {
			utils.LogResponseError(c, h.logger, err)
			return c.JSON(httpErrors.ErrorResponse(err))
		}

		if err = h.authUC.Delete(ctx, uID); err != nil {
			utils.LogResponseError(c, h.logger, err)
			return c.JSON(httpErrors.ErrorResponse(err))
		}

		return c.NoContent(http.StatusOK)
	}
}

func (h *authHandlers) FindByName() echo.HandlerFunc {
	return func(c echo.Context) error {
		span, ctx := opentracing.StartSpanFromContext(utils.GetRequestCtx(c), "authHandlers.FindByName")
		defer span.Finish()

		if c.QueryParam("name") == "" {
			utils.LogResponseError(c, h.logger, httpErrors.NewBadRequestError("name is required"))
			return c.JSON(http.StatusBadRequest, httpErrors.NewBadRequestError("name is required"))
		}

		paginationQuery, err := utils.GetPaginationFromCtx(c)
		if err != nil {
			utils.LogResponseError(c, h.logger, err)
			return c.JSON(httpErrors.ErrorResponse(err))
		}

		response, err := h.authUC.FindByName(ctx, c.QueryParam("name"), paginationQuery)
		if err != nil {
			utils.LogResponseError(c, h.logger, err)
			return c.JSON(httpErrors.ErrorResponse(err))
		}

		return c.JSON(http.StatusOK, response)
	}
}

func (h *authHandlers) GetUsers() echo.HandlerFunc {
	return func(c echo.Context) error {
		span, ctx := opentracing.StartSpanFromContext(utils.GetRequestCtx(c), "authHandlers.GetUsers")
		defer span.Finish()

		paginationQuery, err := utils.GetPaginationFromCtx(c)
		if err != nil {
			utils.LogResponseError(c, h.logger, err)
			return c.JSON(httpErrors.ErrorResponse(err))
		}

		usersList, err := h.authUC.GetUsers(ctx, paginationQuery)
		if err != nil {
			utils.LogResponseError(c, h.logger, err)
			return c.JSON(httpErrors.ErrorResponse(err))
		}

		return c.JSON(http.StatusOK, usersList)
	}
}

func (h *authHandlers) GetMe() echo.HandlerFunc {
	return func(c echo.Context) error {
		span, _ := opentracing.StartSpanFromContext(utils.GetRequestCtx(c), "authHandlers.GetMe")
		defer span.Finish()

		user, ok := c.Get("user").(*models.User)
		if !ok {
			utils.LogResponseError(c, h.logger, httpErrors.NewUnauthorizedError(httpErrors.Unauthorized))
			return utils.ErrResponseWithLog(c, h.logger, httpErrors.NewUnauthorizedError(httpErrors.Unauthorized))
		}

		return c.JSON(http.StatusOK, user)
	}
}

func (h *authHandlers) GetCSRFToken() echo.HandlerFunc {
	return func(c echo.Context) error {
		span, _ := opentracing.StartSpanFromContext(utils.GetRequestCtx(c), "authHandlers.GetCSRFToken")
		defer span.Finish()

		sid, ok := c.Get("sid").(string)
		if !ok {
			utils.LogResponseError(c, h.logger, httpErrors.NewUnauthorizedError(httpErrors.Unauthorized))
			return utils.ErrResponseWithLog(c, h.logger, httpErrors.NewUnauthorizedError(httpErrors.Unauthorized))
		}
		token := csrf.MakeToken(sid, h.logger)
		c.Response().Header().Set(csrf.CSRFHeader, token)
		c.Response().Header().Set("Access-Control-Expose-Headers", csrf.CSRFHeader)

		return c.NoContent(http.StatusOK)
	}
}

func (h *authHandlers) UploadAvatar() echo.HandlerFunc {
	return func(c echo.Context) error {
		span, ctx := opentracing.StartSpanFromContext(utils.GetRequestCtx(c), "authHandlers.UploadAvatar")
		defer span.Finish()

		bucket := c.QueryParam("bucket")
		uID, err := uuid.Parse(c.Param("user_id"))
		if err != nil {
			utils.LogResponseError(c, h.logger, err)
			return c.JSON(httpErrors.ErrorResponse(err))
		}

		image, err := utils.ReadImage(c, "file")
		if err != nil {
			utils.LogResponseError(c, h.logger, err)
			return c.JSON(httpErrors.ErrorResponse(err))
		}

		file, err := image.Open()
		if err != nil {
			utils.LogResponseError(c, h.logger, err)
			return c.JSON(httpErrors.ErrorResponse(err))
		}
		defer file.Close()

		binaryImage := bytes.NewBuffer(nil)
		if _, err = io.Copy(binaryImage, file); err != nil {
			utils.LogResponseError(c, h.logger, err)
			return c.JSON(httpErrors.ErrorResponse(err))
		}

		contentType, err := utils.CheckImageFileContentType(binaryImage.Bytes())
		if err != nil {
			utils.LogResponseError(c, h.logger, err)
			return c.JSON(httpErrors.ErrorResponse(err))
		}

		reader := bytes.NewReader(binaryImage.Bytes())

		updatedUser, err := h.authUC.UploadAvatar(ctx, uID, models.UploadInput{
			File:        reader,
			Name:        image.Filename,
			Size:        image.Size,
			ContentType: contentType,
			BucketName:  bucket,
		})
		if err != nil {
			utils.LogResponseError(c, h.logger, err)
			return c.JSON(httpErrors.ErrorResponse(err))
		}

		return c.JSON(http.StatusOK, updatedUser)
	}
}
