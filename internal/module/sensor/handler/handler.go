package handler

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"sensor-service/config"
	"sensor-service/internal/module/sensor/usecase"
	"sensor-service/internal/platform/httpengine/httpresponse"
)

type Handler struct {
	UseCase usecase.IUseCase
}

func (h Handler) Health(c echo.Context) error {
	return httpresponse.ResponseWithJSON(c, http.StatusOK, "ok")
}

func (h Handler) GetAllSensor(c echo.Context) error {
	request, err := config.MappingRequest(c)
	if err != nil {
		return httpresponse.ResponseWithError(c, http.StatusInternalServerError, err.Error())
	}

	data, errs := h.UseCase.GetAllSensor(request)
	if errs != nil {
		return httpresponse.ResponseWithError(c, http.StatusInternalServerError, httpresponse.ErrorInternalServerError.Message)
	}

	return httpresponse.ResponseWithJSON(c, http.StatusOK, httpresponse.ResponseSuccess(http.StatusOK, "success", data))
}

func (h Handler) UpdateSensor(c echo.Context) error {
	request, err := config.MappingRequest(c)
	if err != nil {
		return httpresponse.ResponseWithError(c, http.StatusInternalServerError, err.Error())
	}

	errs := h.UseCase.UpdateSensor(request)
	if errs != nil {
		return httpresponse.ResponseWithError(c, http.StatusInternalServerError, httpresponse.ErrorInternalServerError.Message)
	}

	return httpresponse.ResponseWithJSON(c, http.StatusOK, httpresponse.ResponseSuccess(http.StatusOK, "success", nil))
}

func (h Handler) DeleteSensor(c echo.Context) error {
	request, err := config.MappingRequest(c)
	if err != nil {
		return httpresponse.ResponseWithError(c, http.StatusInternalServerError, err.Error())
	}

	errs := h.UseCase.DeleteSensor(request)
	if errs != nil {
		return httpresponse.ResponseWithError(c, http.StatusInternalServerError, httpresponse.ErrorInternalServerError.Message)
	}

	return httpresponse.ResponseWithJSON(c, http.StatusOK, httpresponse.ResponseSuccess(http.StatusOK, "success", nil))
}

func NewHandler(useCase usecase.IUseCase) Handler {
	return Handler{
		UseCase: useCase,
	}
}
