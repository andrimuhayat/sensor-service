package config

import (
	"encoding/json"
	"github.com/labstack/echo/v4"
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
