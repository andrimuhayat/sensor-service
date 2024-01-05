// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/api/sensor": {
            "get": {
                "description": "List Sensor Data",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Get List Sensor",
                "parameters": [
                    {
                        "type": "string",
                        "description": "example: [ID1=A, ID2=1], [ID1=B, ID2=2]",
                        "name": "combination_ids",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "example: 01:30AM",
                        "name": "hour_from",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "example: 03:40PM",
                        "name": "hour_to",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "example: 2024-01-02",
                        "name": "date_from",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "example: 2024-01-02",
                        "name": "date_to",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "default": "Bearer \u003cAdd access token here\u003e",
                        "description": "Insert your access token",
                        "name": "Authorization",
                        "in": "header",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/httpresponse.Pagination"
                        }
                    }
                }
            },
            "delete": {
                "description": "Delete Sensor Data",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Delete Sensor",
                "parameters": [
                    {
                        "type": "string",
                        "description": "example: [ID1=A, ID2=1], [ID1=B, ID2=2]",
                        "name": "combination_ids",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "example: 01:30AM",
                        "name": "hour_from",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "example: 03:40PM",
                        "name": "hour_to",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "example: 2024-01-02",
                        "name": "date_from",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "example: 2024-01-02",
                        "name": "date_to",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "default": "Bearer \u003cAdd access token here\u003e",
                        "description": "Insert your access token",
                        "name": "Authorization",
                        "in": "header",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/httpresponse.ResponseHandler"
                        }
                    }
                }
            },
            "patch": {
                "description": "Update Sensor Data",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Update Sensor",
                "parameters": [
                    {
                        "description": "Sensor Value",
                        "name": "sensor_value",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "number"
                        }
                    },
                    {
                        "type": "string",
                        "description": "example: [ID1=A, ID2=1], [ID1=B, ID2=2]",
                        "name": "combination_ids",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "example: 01:30AM",
                        "name": "hour_from",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "example: 03:40PM",
                        "name": "hour_to",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "example: 2024-01-02",
                        "name": "date_from",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "example: 2024-01-02",
                        "name": "date_to",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "default": "Bearer \u003cAdd access token here\u003e",
                        "description": "Insert your access token",
                        "name": "Authorization",
                        "in": "header",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/httpresponse.ResponseHandler"
                        }
                    }
                }
            }
        },
        "/api/user/signin": {
            "post": {
                "description": "Login User",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Sign User",
                "parameters": [
                    {
                        "description": "Sign In",
                        "name": "data",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/dto.Authentication"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/dto.Token"
                        }
                    }
                }
            }
        },
        "/api/user/signup": {
            "post": {
                "description": "signup User",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "signup User",
                "parameters": [
                    {
                        "description": "Sign In",
                        "name": "data",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/entity.User"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/entity.User"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "dto.Authentication": {
            "type": "object",
            "properties": {
                "email": {
                    "type": "string"
                },
                "password": {
                    "type": "string"
                }
            }
        },
        "dto.Token": {
            "type": "object",
            "properties": {
                "email": {
                    "type": "string"
                },
                "role": {
                    "type": "string"
                },
                "token": {
                    "type": "string"
                }
            }
        },
        "entity.User": {
            "type": "object",
            "properties": {
                "email": {
                    "type": "string"
                },
                "password": {
                    "type": "string"
                },
                "roles": {
                    "type": "string"
                }
            }
        },
        "httpresponse.Pagination": {
            "type": "object",
            "properties": {
                "current_page": {
                    "type": "integer"
                },
                "data": {},
                "total_data": {
                    "type": "integer"
                },
                "total_page": {
                    "type": "integer"
                },
                "total_per_page": {
                    "type": "integer"
                }
            }
        },
        "httpresponse.ResponseHandler": {
            "type": "object",
            "properties": {
                "data": {},
                "message": {
                    "type": "string"
                },
                "status_code": {
                    "type": "integer"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "",
	Description:      "",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
