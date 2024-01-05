package middleware

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	"sensor-service/internal/platform/app"
	"sensor-service/internal/platform/httpengine/httpresponse"
	"strings"
)

func AuthorizeJWT(cfg app.App) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auth, err := extractBearerToken(c)
			if err != nil {
				return httpresponse.ResponseWithErrors(c, http.StatusUnauthorized, &httpresponse.HTTPError{
					Code:    http.StatusUnauthorized,
					Message: "UNAUTHORIZED",
				})
			}

			var mySigningKey = []byte(cfg.SecretKey)
			log.Println(cfg.SecretKey)

			token, err := jwt.Parse(*auth, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("There was an error in parsing token.")
				}
				return mySigningKey, nil
			})

			if err != nil {
				return c.JSON(http.StatusUnauthorized, "UNAUTHORIZED")
			}

			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				log.Println(claims)
				c.Set("identity", claims)
				if claims["role"] != "" {
					c.Set("role", claims["role"])
				}
				return next(c)
			}
			log.Println("sfsd?", c.Get("identity"))
			return httpresponse.ResponseWithErrors(c, http.StatusUnauthorized, &httpresponse.HTTPError{
				Code:    http.StatusUnauthorized,
				Message: "UNAUTHORIZED",
			})
		}
	}
}

func extractBearerToken(c echo.Context) (*string, error) {
	authData := c.Request().Header.Get("Authorization")
	if authData == "" {
		return nil, errors.New("authorization can't be nil")
	}
	parts := strings.Split(authData, " ")
	if len(parts) < 2 {
		return nil, errors.New("invalid authorization value")
	}
	if parts[0] != "Bearer" {
		return nil, errors.New("auth should be bearer")
	}

	return &parts[1], nil
}

func CheckRole(role string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Get("role") != role {
				return httpresponse.ResponseWithErrors(c, http.StatusUnauthorized, &httpresponse.HTTPError{
					Code:    http.StatusUnauthorized,
					Message: "UNAUTHORIZED",
				})
			}
			return next(c)
		}
	}
}
