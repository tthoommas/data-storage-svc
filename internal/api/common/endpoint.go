package common

import "github.com/gin-gonic/gin"

type Endpoint interface {
	GetEndpointName() string
	GetEndpointBaseUrl() string
	GetMiddlewares() []gin.HandlerFunc
}

type endpoint struct {
	// Give the endpoint a name (for logs)
	endpointName string
	// The endpoint root url prefix applied to all routes of the endpoint
	endpointBaseUrl string
	// The list of middleware required for the endpoint
	middlewares []gin.HandlerFunc
}

func NewEndpoint(name string, baseUrl string, middlewares []gin.HandlerFunc) Endpoint {
	return endpoint{}
}

func (e endpoint) GetEndpointName() string {
	return e.endpointName
}

func (e endpoint) GetEndpointBaseUrl() string {
	return e.endpointBaseUrl
}

func (e endpoint) GetMiddlewares() []gin.HandlerFunc {
	return e.middlewares
}
