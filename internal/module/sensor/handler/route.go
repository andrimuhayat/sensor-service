package handler

import (
	"github.com/labstack/echo/v4"
	"sensor-service/internal/platform/app"
	"sensor-service/internal/platform/helper"
	"sensor-service/internal/platform/middleware"
)

func NewSensorRoute(h Handler, route *echo.Group, app app.App) {
	sensor := route.Group("/sensor")
	sensor.Use(middleware.AuthorizeJWT(app))
	sensor.GET("", h.GetAllSensor, middleware.CheckRole(helper.Privileges))
	sensor.PATCH("", h.UpdateSensor, middleware.CheckRole(helper.Privileges))
	sensor.DELETE("", h.DeleteSensor, middleware.CheckRole([]string{"admin"}))
	sensor.GET("/health", h.Health)
}
