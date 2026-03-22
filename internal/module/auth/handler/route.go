package handler

import (
	"github.com/labstack/echo/v4"
)

func NewAuthRoute(h Handler, route *echo.Group) {
	sensor := route.Group("/user")
	sensor.POST("/signin", h.SignIn)
	sensor.POST("/signup", h.SignUp)
	sensor.PUT("/changepassword", h.ChangePassword)
	sensor.POST("/removeuser", h.RemoveUser)
}
