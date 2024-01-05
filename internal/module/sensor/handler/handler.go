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

// @Summary Get List Sensor
// @Description List Sensor Data
// @Accept  json
// @Produce  json
// @Param   combination_ids query string false "example: [ID1=A, ID2=1], [ID1=B, ID2=2]"
// @Param   hour_from query string false "example: 01:30AM"
// @Param   hour_to query string false "example: 03:40PM"
// @Param   date_from query string false "example: 2024-01-02"
// @Param   date_to query string false "example: 2024-01-02"
// @Success 200 {object} httpresponse.Pagination
// @Router /api/sensor [get]
// @Param Authorization header string true "Insert your access token" default(Bearer <Add access token here>)
func (h Handler) GetAllSensor(c echo.Context) error {
	request, err := config.MappingRequest(c)
	if err != nil {
		return httpresponse.ResponseWithError(c, http.StatusInternalServerError, err.Error())
	}

	data, errs := h.UseCase.GetAllSensor(request)
	if errs != nil {
		return httpresponse.ResponseWithErrors(c, errs.Code, errs)
	}

	return httpresponse.ResponseWithJSON(c, http.StatusOK, httpresponse.ResponseSuccess(http.StatusOK, "success", data))
}

// @Summary Update Sensor
// @Description Update Sensor Data
// @Accept  json
// @Produce  json
// @Param  sensor_value body number true "Sensor Value"
// @Param   combination_ids query string true "example: [ID1=A, ID2=1], [ID1=B, ID2=2]"
// @Param   hour_from query string true "example: 01:30AM"
// @Param   hour_to query string true "example: 03:40PM"
// @Param   date_from query string true "example: 2024-01-02"
// @Param   date_to query string true "example: 2024-01-02"
// @Success 200 {object} httpresponse.ResponseHandler
// @Router /api/sensor [patch]
// @Param Authorization header string true "Insert your access token" default(Bearer <Add access token here>)
func (h Handler) UpdateSensor(c echo.Context) error {
	request, err := config.MappingRequest(c)
	if err != nil {
		return httpresponse.ResponseWithError(c, http.StatusInternalServerError, err.Error())
	}

	errs := h.UseCase.UpdateSensor(request)
	if errs != nil {
		return httpresponse.ResponseWithErrors(c, errs.Code, errs)
	}

	return httpresponse.ResponseWithJSON(c, http.StatusOK, httpresponse.ResponseSuccess(http.StatusOK, "success", nil))
}

// @Summary Delete Sensor
// @Description Delete Sensor Data
// @Accept  json
// @Produce  json
// @Param   combination_ids query string true "example: [ID1=A, ID2=1], [ID1=B, ID2=2]"
// @Param   hour_from query string true "example: 01:30AM"
// @Param   hour_to query string true "example: 03:40PM"
// @Param   date_from query string true "example: 2024-01-02"
// @Param   date_to query string true "example: 2024-01-02"
// @Success 200 {object} httpresponse.ResponseHandler
// @Router /api/sensor [delete]
// @Param Authorization header string true "Insert your access token" default(Bearer <Add access token here>)
func (h Handler) DeleteSensor(c echo.Context) error {
	request, err := config.MappingRequest(c)
	if err != nil {
		return httpresponse.ResponseWithError(c, http.StatusInternalServerError, err.Error())
	}

	errs := h.UseCase.DeleteSensor(request)
	if errs != nil {
		return httpresponse.ResponseWithErrors(c, errs.Code, errs)
	}

	return httpresponse.ResponseWithJSON(c, http.StatusOK, httpresponse.ResponseSuccess(http.StatusOK, "success", nil))
}

func CheckAdminRole(c echo.Context) error {
	if c.Get("role") != "admin" {
		return httpresponse.ResponseWithError(c, http.StatusUnauthorized, "UNAUTHORIZED")
	}
	return nil
}

func NewHandler(useCase usecase.IUseCase) Handler {
	return Handler{
		UseCase: useCase,
	}
}
