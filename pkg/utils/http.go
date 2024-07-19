package utils

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/GDLMNV/api-mc/config"
	"github.com/GDLMNV/api-mc/internal/models"
	"github.com/GDLMNV/api-mc/pkg/httpErrors"
	"github.com/GDLMNV/api-mc/pkg/logger"
	"github.com/GDLMNV/api-mc/pkg/sanitize"
)

func GetRequestID(c echo.Context) string {
	return c.Response().Header().Get(echo.HeaderXRequestID)
}

// ReqIDCtxKey is a key used for the Request ID in context
type ReqIDCtxKey struct{}

func GetCtxWithReqID(c echo.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(c.Request().Context(), time.Second*15)
	ctx = context.WithValue(ctx, ReqIDCtxKey{}, GetRequestID(c))
	return ctx, cancel
}

func GetRequestCtx(c echo.Context) context.Context {
	return context.WithValue(c.Request().Context(), ReqIDCtxKey{}, GetRequestID(c))
}

func GetConfigPath(configPath string) string {
	if configPath == "docker" {
		return "./config/config-docker"
	}
	return "./config/config-local"
}

func ConfigureJWTCookie(cfg *config.Config, jwtToken string) *http.Cookie {
	return &http.Cookie{
		Name:       cfg.Cookie.Name,
		Value:      jwtToken,
		Path:       "/",
		RawExpires: "",
		MaxAge:     cfg.Cookie.MaxAge,
		Secure:     cfg.Cookie.Secure,
		HttpOnly:   cfg.Cookie.HTTPOnly,
		SameSite:   0,
	}
}

func CreateSessionCookie(cfg *config.Config, session string) *http.Cookie {
	return &http.Cookie{
		Name:  cfg.Session.Name,
		Value: session,
		Path:  "/",
		// Domain: "/",
		// Expires:    time.Now().Add(1 * time.Minute),
		RawExpires: "",
		MaxAge:     cfg.Session.Expire,
		Secure:     cfg.Cookie.Secure,
		HttpOnly:   cfg.Cookie.HTTPOnly,
		SameSite:   0,
	}
}

func DeleteSessionCookie(c echo.Context, sessionName string) {
	c.SetCookie(&http.Cookie{
		Name:   sessionName,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
}

type UserCtxKey struct{}

func GetUserFromCtx(ctx context.Context) (*models.User, error) {
	user, ok := ctx.Value(UserCtxKey{}).(*models.User)
	if !ok {
		return nil, httpErrors.Unauthorized
	}

	return user, nil
}

func GetIPAddress(c echo.Context) string {
	return c.Request().RemoteAddr
}

func ErrResponseWithLog(ctx echo.Context, logger logger.Logger, err error) error {
	logger.Errorf(
		"ErrResponseWithLog, RequestID: %s, IPAddress: %s, Error: %s",
		GetRequestID(ctx),
		GetIPAddress(ctx),
		err,
	)
	return ctx.JSON(httpErrors.ErrorResponse(err))
}

func LogResponseError(ctx echo.Context, logger logger.Logger, err error) {
	logger.Errorf(
		"ErrResponseWithLog, RequestID: %s, IPAddress: %s, Error: %s",
		GetRequestID(ctx),
		GetIPAddress(ctx),
		err,
	)
}

func ReadRequest(ctx echo.Context, request interface{}) error {
	if err := ctx.Bind(request); err != nil {
		return err
	}
	return validate.StructCtx(ctx.Request().Context(), request)
}

func ReadImage(ctx echo.Context, field string) (*multipart.FileHeader, error) {
	image, err := ctx.FormFile(field)
	if err != nil {
		return nil, errors.WithMessage(err, "ctx.FormFile")
	}

	if err = CheckImageContentType(image); err != nil {
		return nil, err
	}

	return image, nil
}

func SanitizeRequest(ctx echo.Context, request interface{}) error {
	body, err := ioutil.ReadAll(ctx.Request().Body)
	if err != nil {
		return err
	}
	defer ctx.Request().Body.Close()

	sanBody, err := sanitize.SanitizeJSON(body)
	if err != nil {
		return ctx.NoContent(http.StatusBadRequest)
	}

	if err = json.Unmarshal(sanBody, request); err != nil {
		return err
	}

	return validate.StructCtx(ctx.Request().Context(), request)
}

var allowedImagesContentTypes = map[string]string{
	"image/bmp":                "bmp",
	"image/gif":                "gif",
	"image/png":                "png",
	"image/jpeg":               "jpeg",
	"image/jpg":                "jpg",
	"image/svg+xml":            "svg",
	"image/webp":               "webp",
	"image/tiff":               "tiff",
	"image/vnd.microsoft.icon": "ico",
}

func CheckImageFileContentType(fileContent []byte) (string, error) {
	contentType := http.DetectContentType(fileContent)

	extension, ok := allowedImagesContentTypes[contentType]
	if !ok {
		return "", errors.New("this content type is not allowed")
	}

	return extension, nil
}
