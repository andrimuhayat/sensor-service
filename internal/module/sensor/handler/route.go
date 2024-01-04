package handler

import "github.com/labstack/echo/v4"

func NewSensorRoute(h Handler, route *echo.Group) {
	//sensor := route.Group("")
	//sensor.POST("", h.CreateUser)
	route.GET("/", h.Health)
	//sensor.PUT("/update/status", h.UpdateUserStatus, middleware.AuthorizeJWT(cfg))
}
