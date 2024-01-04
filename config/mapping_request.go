package config

import (
	"encoding/json"
	"github.com/labstack/echo/v4"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

type Result struct {
	httpStatus int
	code       int
	message    string
	data       interface{}
	isFailure  bool
	retry      bool
}

type HTTPRequest struct {
	Queries    map[string]interface{}
	RawQueries map[string][]string
	Headers    map[string]string
	Params     map[string]string
	Body       map[string]interface{}
	formData   map[string]string
}

func (H HTTPRequest) Request() *http.Request {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) SetRequest(r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) SetResponse(r *echo.Response) {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) Response() *echo.Response {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) IsTLS() bool {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) IsWebSocket() bool {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) Scheme() string {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) RealIP() string {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) Path() string {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) SetPath(p string) {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) Param(name string) string {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) ParamNames() []string {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) SetParamNames(names ...string) {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) ParamValues() []string {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) SetParamValues(values ...string) {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) QueryParam(name string) string {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) QueryParams() url.Values {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) QueryString() string {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) FormValue(name string) string {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) FormParams() (url.Values, error) {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) FormFile(name string) (*multipart.FileHeader, error) {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) MultipartForm() (*multipart.Form, error) {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) Cookie(name string) (*http.Cookie, error) {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) SetCookie(cookie *http.Cookie) {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) Cookies() []*http.Cookie {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) Get(key string) interface{} {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) Set(key string, val interface{}) {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) Bind(i interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) Validate(i interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) Render(code int, name string, data interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) HTML(code int, html string) error {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) HTMLBlob(code int, b []byte) error {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) String(code int, s string) error {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) JSON(code int, i interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) JSONPretty(code int, i interface{}, indent string) error {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) JSONBlob(code int, b []byte) error {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) JSONP(code int, callback string, i interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) JSONPBlob(code int, callback string, b []byte) error {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) XML(code int, i interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) XMLPretty(code int, i interface{}, indent string) error {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) XMLBlob(code int, b []byte) error {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) Blob(code int, contentType string, b []byte) error {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) Stream(code int, contentType string, r io.Reader) error {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) File(file string) error {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) Attachment(file string, name string) error {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) Inline(file string, name string) error {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) NoContent(code int) error {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) Redirect(code int, url string) error {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) Error(err error) {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) Handler() echo.HandlerFunc {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) SetHandler(h echo.HandlerFunc) {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) Logger() echo.Logger {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) SetLogger(l echo.Logger) {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) Echo() *echo.Echo {
	//TODO implement me
	panic("implement me")
}

func (H HTTPRequest) Reset(r *http.Request, w http.ResponseWriter) {
	//TODO implement me
	panic("implement me")
}

func MappingRequest(req echo.Context) (HTTPRequest, error) {
	var (
		headers  = make(map[string]string)
		queries  = make(map[string]interface{})
		params   = make(map[string]string)
		formData = make(map[string]string)
		body     = make(map[string]interface{})
	)

	err := json.NewDecoder(req.Request().Body).Decode(&body)
	if err != nil {
	}

	for key, values := range req.Request().Header {
		for _, value := range values {
			headers[key] = value
		}
	}

	form, _ := req.MultipartForm()

	if form != nil {
		for k, v := range form.Value {
			formData[k] = v[0]
		}
	}

	for key, values := range req.QueryParams() {
		for _, value := range values {
			queries[key] = value
		}
	}
	route := req.Path()

	// Extract the parameter keys from the route path
	routeParams := extractParamKeys(route)

	for _, paramKey := range routeParams {
		value := req.Param(paramKey)
		params[paramKey] = value
	}

	return HTTPRequest{
		Params:  params,
		Queries: queries,
		Body:    body,
		Headers: headers,
	}, nil
}

// Extract parameter keys from the route path
func extractParamKeys(routePath string) []string {
	keys := make([]string, 0)

	segments := strings.Split(routePath, "/")
	for _, segment := range segments {
		if strings.HasPrefix(segment, ":") {
			paramKey := strings.TrimPrefix(segment, ":")
			keys = append(keys, paramKey)
		}
	}

	return keys
}
