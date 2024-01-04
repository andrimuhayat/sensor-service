package handler

import "github.com/labstack/echo/v4"

func NewSensorRoute(h Handler, route *echo.Group) {
	sensor := route.Group("/sensor")
	sensor.GET("", h.GetAllSensor)
	sensor.PATCH("", h.UpdateSensor)
	sensor.DELETE("", h.DeleteSensor)
	sensor.GET("/health", h.Health)
	//sensor.PUT("/update/status", h.UpdateUserStatus, middleware.AuthorizeJWT(cfg))
}
