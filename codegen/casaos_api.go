// Package codegen provides types generated from the CasaOS OpenAPI specification.
// This file was generated from api/casaos/openapi.yaml
package codegen

import (
	"encoding/json"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
)

// BaseResponse defines model for BaseResponse.
type BaseResponse struct {
	// Message returned by server side if there is any
	Message *string `json:"message,omitempty"`
}

// SuccessResponseString defines model for SuccessResponseString.
type SuccessResponseString struct {
	// Message returned by server side if there is any
	Message *string `json:"message,omitempty"`
	// Data is the specific success information
	Data *string `json:"data,omitempty"`
}

// HealthServices defines model for HealthServices.
type HealthServices struct {
	Running    *[]string `json:"running,omitempty"`
	NotRunning *[]string `json:"not_running,omitempty"`
}

// HealthPorts defines model for HealthPorts.
type HealthPorts struct {
	TCP *[]int `json:"tcp,omitempty"`
	UDP *[]int `json:"udp,omitempty"`
}

// ZTInfo defines model for ZTInfo.
type ZTInfo struct {
	Id     *string `json:"id,omitempty"`
	Name   *string `json:"name,omitempty"`
	Status *string `json:"status,omitempty"`
}

// ResponseOK defines model for ResponseOK.
type ResponseOK = BaseResponse

// ResponseInternalServerError defines model for ResponseInternalServerError.
type ResponseInternalServerError = BaseResponse

// ResponseClientError defines model for ResponseClientError.
type ResponseClientError = BaseResponse

// GetHealthServicesOK defines model for GetHealthServicesOK.
type GetHealthServicesOK struct {
	Message *string         `json:"message,omitempty"`
	Data    *HealthServices `json:"data,omitempty"`
}

// GetHealthPortsOK defines model for GetHealthPortsOK.
type GetHealthPortsOK struct {
	Message *string      `json:"message,omitempty"`
	Data    *HealthPorts `json:"data,omitempty"`
}

// GetZTInfoOK defines model for GetZTInfoOK.
type GetZTInfoOK = ZTInfo

// ResponseStringOK defines model for ResponseStringOK.
type ResponseStringOK = SuccessResponseString

// CheckUploadChunkParams defines parameters for CheckUploadChunk.
type CheckUploadChunkParams struct {
	// Path is the file path
	Path string `json:"path"`
	// RelativePath is the relative file path
	RelativePath string `json:"relativePath"`
	// Filename is the file name
	Filename string `json:"filename"`
	// ChunkNumber is the chunk number
	ChunkNumber string `json:"chunkNumber"`
	// TotalChunks is the total chunks
	TotalChunks int `json:"totalChunks"`
}

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Get service status
	// (GET /health/services)
	GetHealthServices(ctx echo.Context) error
	// Get port in use
	// (GET /health/ports)
	GetHealthPorts(ctx echo.Context) error
	// Get log
	// (GET /health/logs)
	GetHealthlogs(ctx echo.Context) error
	// Test file methods
	// (GET /file/test)
	GetFileTest(ctx echo.Context) error
	// Check upload chunk
	// (GET /file/upload)
	CheckUploadChunk(ctx echo.Context, params CheckUploadChunkParams) error
	// Upload file
	// (POST /file/upload)
	PostUploadFile(ctx echo.Context) error
	// Get Zerotier info
	// (GET /zt/info)
	GetZerotierInfo(ctx echo.Context) error
	// Set Zerotier network status
	// (PUT /zt/{network_id}/status)
	SetZerotierNetworkStatus(ctx echo.Context, networkId string) error
}

// EchoRouter is the interface for echo route registration
type EchoRouter interface {
	CONNECT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	HEAD(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	TRACE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
}

// RegisterHandlers adds each server route to the EchoRouter.
func RegisterHandlers(router EchoRouter, si ServerInterface) {
	RegisterHandlersWithBaseURL(router, si, "")
}

// RegisterHandlersWithBaseURL registers handlers with a base URL prefix.
func RegisterHandlersWithBaseURL(router EchoRouter, si ServerInterface, baseURL string) {
	router.GET(baseURL+"/health/services", si.GetHealthServices)
	router.GET(baseURL+"/health/ports", si.GetHealthPorts)
	router.GET(baseURL+"/health/logs", si.GetHealthlogs)
	router.GET(baseURL+"/file/test", si.GetFileTest)
	router.GET(baseURL+"/file/upload", func(ctx echo.Context) error {
		var params CheckUploadChunkParams
		params.Path = ctx.QueryParam("path")
		params.RelativePath = ctx.QueryParam("relativePath")
		params.Filename = ctx.QueryParam("filename")
		params.ChunkNumber = ctx.QueryParam("chunkNumber")
		return si.CheckUploadChunk(ctx, params)
	})
	router.POST(baseURL+"/file/upload", si.PostUploadFile)
	router.GET(baseURL+"/zt/info", si.GetZerotierInfo)
	router.PUT(baseURL+"/zt/:network_id/status", func(ctx echo.Context) error {
		networkId := ctx.Param("network_id")
		return si.SetZerotierNetworkStatus(ctx, networkId)
	})
}

// swaggerSpec is the parsed OpenAPI spec
var swaggerSpec *openapi3.T

// GetSwagger returns the OpenAPI specification.
func GetSwagger() (*openapi3.T, error) {
	if swaggerSpec != nil {
		return swaggerSpec, nil
	}

	// Parse the spec from the embedded YAML
	loader := openapi3.NewLoader()
	spec, err := loader.LoadFromFile("api/casaos/openapi.yaml")
	if err != nil {
		// Fallback: return a minimal valid spec
		minSpec := []byte(`{"openapi":"3.0.3","info":{"title":"CasaOS API","version":"v2"},"paths":{}}`)
		spec = &openapi3.T{}
		if jsonErr := json.Unmarshal(minSpec, spec); jsonErr != nil {
			return nil, err
		}
	}

	// Ensure valid servers slice
	if len(spec.Servers) == 0 {
		spec.Servers = openapi3.Servers{&openapi3.Server{URL: "/v2/casaos"}}
	}

	swaggerSpec = spec
	return spec, nil
}

// Handler wraps ServerInterface with a base URL
type Handler struct {
	http.Handler
}
