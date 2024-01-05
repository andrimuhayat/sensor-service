package handler

import (
	"github.com/labstack/echo/v4"
	"sensor-service/internal/platform/app"
	"sensor-service/internal/platform/middleware"
)

func NewSensorRoute(h Handler, route *echo.Group, app app.App) {
	sensor := route.Group("/sensor")
	sensor.Use(middleware.AuthorizeJWT(app))
	sensor.GET("", h.GetAllSensor, middleware.CheckRole("admin"))
	sensor.PATCH("", h.UpdateSensor)
	sensor.DELETE("", h.DeleteSensor)
	sensor.GET("/health", h.Health)
}
