package httpresponse

import "fmt"

type HTTPError struct {
	Code    int    `json:"status_code"`
	Message string `json:"message"`
}

func (error *HTTPError) Error() string {
	return error.Message
}

func (error *HTTPError) StatusCode() int {
	return error.Code
}

var (
	ErrorDisbursementsPaymentTpe = HTTPError{Message: "Cannot use OVO or go-pay for disbursements"}
	ErrorInternalServerError     = HTTPError{Message: "Internal Server Error"}
	ErrorBadRequest              = HTTPError{Message: "Bad Request"}
	ErrorPasswordNotMatch        = HTTPError{Message: "Password doesnt match"}
	ErrorSameLastPassword        = HTTPError{Message: "Your password is the same as one of your last 5 passwords. Please choose a different password."}
)

func DataNotFound(field string) string {
	return fmt.Sprintf(`%sDoesn't Exists'`, field)
}

func DataExists(field string) string {
	return fmt.Sprintf(`%sAlready Exists`, field)
}
