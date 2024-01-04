package httpresponse

import (
	"github.com/labstack/echo/v4"
	"log"
)

type ResponseHandler struct {
	StatusCode int         `json:"status_code"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data"`
}

type Pagination struct {
	Data         interface{} `json:"data"`
	TotalData    int         `json:"total_data"`
	TotalPage    int         `json:"total_page"`
	CurrentPage  int         `json:"current_page"`
	TotalPerPage int         `json:"total_per_page"`
}

func ResponseSuccess(status int, message string, data interface{}) ResponseHandler {
	return ResponseHandler{
		StatusCode: status,
		Message:    message,
		Data:       data,
	}
}

func ResponseWithJSON(c echo.Context, code int, payload interface{}) error {
	//response, _ := json.Marshal(payload)
	c.Set("Content-Type", "application/json")
	return c.JSON(code, payload)
}

func ResponseWithErrors(c echo.Context, code int, msg *HTTPError) error {
	return ResponseWithJSON(c, code, msg)
}

func ResponseWithError(c echo.Context, code int, msg string) error {
	log.Println("code", code)
	return ResponseWithJSON(c, code, map[string]string{"message": msg})
}
