package httpengine

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

// Router is the interface to be implemented by the HTTP routers
type Router interface {
	GET(uri string, f func(echo.Context, *http.Request))
	POST(uri string, f func(echo.Context, *http.Request))

	SERVE(port string)
	GetRouter() *echo.Echo
}
