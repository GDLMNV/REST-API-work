package news

import "github.com/labstack/echo/v4"

type Handlers interface {
	Create() echo.HandlerFunc
	Update() echo.HandlerFunc
	GetByID() echo.HandlerFunc
	Delete() echo.HandlerFunc
	GetNews() echo.HandlerFunc
	SearchByTitle() echo.HandlerFunc
}
