package handler

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"sensor-service/config"
	forgotpassword "sensor-service/internal/module/forgot-password/usecase"
	"sensor-service/internal/platform/httpengine/httpresponse"
)

type Handler struct {
	UseCase forgotpassword.IUseCase
}

// @Summary Initiate Forgot Password
// @Description Generate reset token for password recovery
// @Accept  json
// @Produce  json
// @Param data body map[string]string true "Email"
// @Success 200 {object} map[string]string
// @Router /api/forgot-password/initiate [post]
func (h Handler) InitiateForgotPassword(c echo.Context) error {
	request, err := config.MappingRequest(c)
	if err != nil {
		return httpresponse.ResponseWithError(c, http.StatusInternalServerError, err.Error())
	}

	email, ok := request.Body["email"].(string)
	if !ok || email == "" {
		return httpresponse.ResponseWithError(c, http.StatusBadRequest, "Email is required")
	}

	resetToken, errs := h.UseCase.InitiateForgotPassword(email)
	if errs != nil {
		return httpresponse.ResponseWithErrors(c, errs.Code, errs)
	}

	data := map[string]string{
		"resetToken": resetToken,
	}

	return httpresponse.ResponseWithJSON(c, http.StatusOK, httpresponse.ResponseSuccess(http.StatusOK, "success", data))
}

// @Summary Validate Reset Token
// @Description Validate if reset token is valid and not expired
// @Accept  json
// @Produce  json
// @Param data body map[string]string true "Reset Token"
// @Success 200 {object} map[string]string
// @Router /api/forgot-password/validate [post]
func (h Handler) ValidateResetToken(c echo.Context) error {
	request, err := config.MappingRequest(c)
	if err != nil {
		return httpresponse.ResponseWithError(c, http.StatusInternalServerError, err.Error())
	}

	resetToken, ok := request.Body["resetToken"].(string)
	if !ok || resetToken == "" {
		return httpresponse.ResponseWithError(c, http.StatusBadRequest, "Reset token is required")
	}

	email, errs := h.UseCase.ValidateResetToken(resetToken)
	if errs != nil {
		return httpresponse.ResponseWithErrors(c, errs.Code, errs)
	}

	data := map[string]string{
		"email": email,
	}

	return httpresponse.ResponseWithJSON(c, http.StatusOK, httpresponse.ResponseSuccess(http.StatusOK, "success", data))
}

// @Summary Complete Password Reset
// @Description Update password using valid reset token
// @Accept  json
// @Produce  json
// @Param data body map[string]string true "Reset Token and New Password"
// @Success 200 {object} map[string]string
// @Router /api/forgot-password/complete [post]
func (h Handler) CompleteForgotPassword(c echo.Context) error {
	request, err := config.MappingRequest(c)
	if err != nil {
		return httpresponse.ResponseWithError(c, http.StatusInternalServerError, err.Error())
	}

	resetToken, ok := request.Body["resetToken"].(string)
	if !ok || resetToken == "" {
		return httpresponse.ResponseWithError(c, http.StatusBadRequest, "Reset token is required")
	}

	newPassword, ok := request.Body["newPassword"].(string)
	if !ok || newPassword == "" {
		return httpresponse.ResponseWithError(c, http.StatusBadRequest, "New password is required")
	}

	errs := h.UseCase.CompleteForgotPassword(resetToken, newPassword)
	if errs != nil {
		return httpresponse.ResponseWithErrors(c, errs.Code, errs)
	}

	data := map[string]string{
		"message": "Password reset successfully",
	}

	return httpresponse.ResponseWithJSON(c, http.StatusOK, httpresponse.ResponseSuccess(http.StatusOK, "success", data))
}

func NewHandler(useCase forgotpassword.IUseCase) Handler {
	return Handler{
		UseCase: useCase,
	}
}
