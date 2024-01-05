package usecase

import (
	"github.com/mitchellh/mapstructure"
	"golang.org/x/exp/slices"
	"log"
	"net/http"
	"sensor-service/config"
	"sensor-service/internal/module/auth/dto"
	"sensor-service/internal/module/auth/entity"
	auth "sensor-service/internal/module/auth/repository"
	"sensor-service/internal/platform/app"
	"sensor-service/internal/platform/helper"
	"sensor-service/internal/platform/httpengine/httpresponse"
)

type IUseCase interface {
	SignIn(request config.HTTPRequest) (dto.Token, *httpresponse.HTTPError)
	SignUp(request config.HTTPRequest) (entity.User, *httpresponse.HTTPError)
}

type UseCase struct {
	GenericRepository auth.IGenericRepository
	AppCfg            app.App
}

func (u UseCase) SignIn(request config.HTTPRequest) (dto.Token, *httpresponse.HTTPError) {
	var err error
	httpError := httpresponse.HTTPError{}

	var requestAuth dto.Authentication

	config := helper.DecoderConfig(&requestAuth)
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		httpError.Code = http.StatusInternalServerError
		httpError.Message = httpresponse.ErrorInternalServerError.Message
		return dto.Token{}, &httpError
	}
	if err = decoder.Decode(request.Body); err != nil {
		log.Println("{SignIn}{Decode}{Error} : ", err)
	}

	authUser, err := u.GenericRepository.FindByEmail(entity.User{}, requestAuth.Email)
	if err != nil {
		log.Println("{SignIn}{FindByEmail}{Error} : ", err)
		httpError.Code = http.StatusInternalServerError
		httpError.Message = httpresponse.ErrorInternalServerError.Message
		return dto.Token{}, &httpError
	}

	if authUser == nil {
		httpError.Code = http.StatusBadRequest
		httpError.Message = "Username or Password is incorrect"
		return dto.Token{}, &httpError
	}

	user, _ := helper.TypeConverter[entity.User](&authUser)

	check := helper.CheckPasswordHash(requestAuth.Password, user.Password)

	if !check {
		httpError.Code = http.StatusBadRequest
		httpError.Message = "Username or Password is incorrect"
		return dto.Token{}, &httpError
	}

	validToken, err := helper.GenerateJWT(user.Email, user.Role, u.AppCfg.SecretKey)
	if err != nil {
		httpError.Code = http.StatusBadRequest
		httpError.Message = "Failed to generate token"
		return dto.Token{}, &httpError
	}

	var token dto.Token
	token.Email = user.Email
	token.Role = user.Role
	token.TokenString = validToken

	return token, nil
}

func (u UseCase) SignUp(request config.HTTPRequest) (entity.User, *httpresponse.HTTPError) {
	var err error
	httpError := httpresponse.HTTPError{}

	var user entity.User

	config := helper.DecoderConfig(&user)
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		httpError.Code = http.StatusInternalServerError
		httpError.Message = httpresponse.ErrorInternalServerError.Message
		return entity.User{}, &httpError
	}
	if err = decoder.Decode(request.Body); err != nil {
		log.Println("{SignIn}{Decode}{Error} : ", err)
	}

	if !slices.Contains(helper.Privileges, user.Role) {
		httpError.Code = http.StatusBadRequest
		httpError.Message = "Role doesn't exists, please use admin or user"
		return entity.User{}, &httpError
	}

	authUser, err := u.GenericRepository.FindByEmail(entity.User{}, user.Email)
	if err != nil {
		httpError.Code = http.StatusInternalServerError
		httpError.Message = httpresponse.ErrorInternalServerError.Message
		return entity.User{}, &httpError
	}

	if authUser != nil {
		httpError.Code = http.StatusBadRequest
		httpError.Message = "email cannot be same"
		return entity.User{}, &httpError
	}
	user.Password, err = helper.GeneratehashPassword(user.Password)
	if err != nil {
		log.Fatalln("Error in password hashing.")
	}

	err = u.GenericRepository.Create(user)
	if err != nil {
		httpError.Code = http.StatusInternalServerError
		httpError.Message = httpresponse.ErrorInternalServerError.Message
		return entity.User{}, &httpError
	}
	return user, nil
}

func NewUseCase(genericRepository auth.IGenericRepository, app app.App) IUseCase {
	return UseCase{
		GenericRepository: genericRepository,
		AppCfg:            app,
	}
}
