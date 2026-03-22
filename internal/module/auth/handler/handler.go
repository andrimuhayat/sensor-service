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

// @Summary Change Password
// @Description Change user password (requires authentication)
// @Accept  json
// @Produce  json
// @Param data body dto.ChangePasswordRequest true "Change Password"
// @Success 200 {object} httpresponse.ResponseHandler
// @Router /api/user/changepassword [put]
// @Param Authorization header string true "Insert your access token" default(Bearer <Add access token here>)
func (h Handler) ChangePassword(c echo.Context) error {
	request, err := config.MappingRequest(c)
	if err != nil {
		return httpresponse.ResponseWithError(c, http.StatusInternalServerError, err.Error())
	}

	var changePwdReq struct {
		Email       string `json:"email"`
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}

	if err := config.MapToStruct(request.Body, &changePwdReq); err != nil {
		return httpresponse.ResponseWithError(c, http.StatusBadRequest, err.Error())
	}

	errs := h.UseCase.ChangePassword(changePwdReq.Email, changePwdReq.OldPassword, changePwdReq.NewPassword)
	if errs != nil {
		return httpresponse.ResponseWithErrors(c, errs.Code, errs)
	}

	return httpresponse.ResponseWithJSON(c, http.StatusOK, httpresponse.ResponseSuccess(http.StatusOK, "success", nil))
}

// @Summary Remove User
// @Description Remove user by email (admin only)
// @Accept  json
// @Produce  json
// @Param data body RemoveUserRequest true "Remove User"
// @Success 200 {object} httpresponse.ResponseHandler
// @Router /api/user/removeuser [post]
// @Param Authorization header string true "Insert your access token" default(Bearer <Add access token here>)
func (h Handler) RemoveUser(c echo.Context) error {
	request, err := config.MappingRequest(c)
	if err != nil {
		return httpresponse.ResponseWithError(c, http.StatusInternalServerError, err.Error())
	}

	var removeUserReq struct {
		Email string `json:"email"`
	}

	if err := config.MapToStruct(request.Body, &removeUserReq); err != nil {
		return httpresponse.ResponseWithError(c, http.StatusBadRequest, err.Error())
	}

	// Extract caller role from context (set by auth middleware after JWT validation)
	callerRole, ok := c.Get("role").(string)
	if !ok {
		return httpresponse.ResponseWithError(c, http.StatusUnauthorized, "Unauthorized: missing role in context")
	}

	errs := h.UseCase.RemoveUser(removeUserReq.Email, callerRole)
	if errs != nil {
		return httpresponse.ResponseWithErrors(c, errs.Code, errs)
	}

	return httpresponse.ResponseWithJSON(c, http.StatusOK, httpresponse.ResponseSuccess(http.StatusOK, "success", nil))
}

func NewHandler(useCase auth.IUseCase) Handler {
	return Handler{
		UseCase: useCase,
	}
}