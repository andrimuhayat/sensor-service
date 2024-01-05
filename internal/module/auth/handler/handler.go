package handler

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"sensor-service/config"
	auth "sensor-service/internal/module/auth/usecase"
	"sensor-service/internal/platform/httpengine/httpresponse"
)

type Handler struct {
	UseCase auth.IUseCase
}

// @Summary Sign User
// @Description Login User
// @Accept  json
// @Produce  json
// @Param data body dto.Authentication true "Sign In"
// @Success 200 {object} dto.Token
// @Router /api/user/signin [post]
func (h Handler) SignIn(c echo.Context) error {
	request, err := config.MappingRequest(c)
	if err != nil {
		return httpresponse.ResponseWithError(c, http.StatusInternalServerError, err.Error())
	}

	data, errs := h.UseCase.SignIn(request)
	if errs != nil {
		return httpresponse.ResponseWithErrors(c, errs.Code, errs)
	}

	return httpresponse.ResponseWithJSON(c, http.StatusOK, httpresponse.ResponseSuccess(http.StatusOK, "success", data))
}

// @Summary signup User
// @Description signup User
// @Accept  json
// @Produce  json
// @Param data body entity.User true "Sign In"
// @Success 200 {object} entity.User
// @Router /api/user/signup [post]
func (h Handler) SignUp(c echo.Context) error {
	request, err := config.MappingRequest(c)
	if err != nil {
		return httpresponse.ResponseWithError(c, http.StatusInternalServerError, err.Error())
	}

	data, errs := h.UseCase.SignUp(request)
	if errs != nil {
		return httpresponse.ResponseWithErrors(c, errs.Code, errs)
	}

	return httpresponse.ResponseWithJSON(c, http.StatusOK, httpresponse.ResponseSuccess(http.StatusOK, "success", data))
}

func NewHandler(useCase auth.IUseCase) Handler {
	return Handler{
		UseCase: useCase,
	}
}
