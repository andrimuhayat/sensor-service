package middleware

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"runtime"
	"sensor-service/internal/platform/httpengine/httpresponse"
)

func PanicException(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		defer func() {
			if err := recover(); err != nil {
				buf := make([]byte, 2048)
				n := runtime.Stack(buf, false)
				buf = buf[:n]
				fmt.Printf("recovering from err %v\n %s", err, buf)
				httpresponse.ResponseWithError(c, http.StatusInternalServerError, httpresponse.ErrorInternalServerError.Message)
			}
		}()
		return next(c)
	}
}
