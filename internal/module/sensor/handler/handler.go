package handler

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"sensor-service/internal/module/sensor/usecase"
	"sensor-service/internal/platform/httpengine/httpresponse"
)

type Handler struct {
	UseCase usecase.IUseCase
}

func (h Handler) Health(c echo.Context) error {
	return httpresponse.ResponseWithJSON(c, http.StatusOK, "ok")
}

func NewHandler(useCase usecase.IUseCase) Handler {
	return Handler{
		UseCase: useCase,
	}
}
