package common

import "github.com/gin-gonic/gin"

type EndpointGroup interface {
	GetEndpointName() string
	GetGroupUrl() string
	GetCommonMiddlewares() []gin.HandlerFunc
	GetPermissionsManager() PermissionsManager
	GetEndpointsList() map[MethodPath][]gin.HandlerFunc
}

type endpointGroup struct {
	// Give the endpoint a name (for logs)
	endpointName string
	// The endpoint root url prefix applied to all routes of the endpoint
	groupUrl string
	// The list of middleware required for the endpoint
	commonMiddlewares []gin.HandlerFunc
	// The list of http endpoints exposed, key is the http method to use
	endpoints map[MethodPath][]gin.HandlerFunc
	// The permissions manager injected in all endpoints
	permissionsManager PermissionsManager
}

func NewEndpoint(name string, groupUrl string, commonMiddlewares []gin.HandlerFunc, endpoints map[MethodPath][]gin.HandlerFunc, permissionManager PermissionsManager) EndpointGroup {
	return endpointGroup{
		endpointName:       name,
		groupUrl:           groupUrl,
		commonMiddlewares:  commonMiddlewares,
		endpoints:          endpoints,
		permissionsManager: permissionManager,
	}
}

func (e endpointGroup) GetEndpointName() string {
	return e.endpointName
}

func (e endpointGroup) GetGroupUrl() string {
	return e.groupUrl
}

func (e endpointGroup) GetCommonMiddlewares() []gin.HandlerFunc {
	return e.commonMiddlewares
}

func (e endpointGroup) GetEndpointsList() map[MethodPath][]gin.HandlerFunc {
	return e.endpoints
}

func (e endpointGroup) GetPermissionsManager() PermissionsManager {
	return e.permissionsManager
}

type MethodPath struct {
	Method string
	Path   string
}
