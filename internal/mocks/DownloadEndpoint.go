// Code generated by mockery v2.53.3. DO NOT EDIT.

package mocks

import (
	common "data-storage-svc/internal/api/common"

	gin "github.com/gin-gonic/gin"

	mock "github.com/stretchr/testify/mock"
)

// DownloadEndpoint is an autogenerated mock type for the DownloadEndpoint type
type DownloadEndpoint struct {
	mock.Mock
}

// Download provides a mock function with given fields: c
func (_m *DownloadEndpoint) Download(c *gin.Context) {
	_m.Called(c)
}

// Get provides a mock function with given fields: c
func (_m *DownloadEndpoint) Get(c *gin.Context) {
	_m.Called(c)
}

// GetCommonMiddlewares provides a mock function with no fields
func (_m *DownloadEndpoint) GetCommonMiddlewares() []gin.HandlerFunc {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetCommonMiddlewares")
	}

	var r0 []gin.HandlerFunc
	if rf, ok := ret.Get(0).(func() []gin.HandlerFunc); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]gin.HandlerFunc)
		}
	}

	return r0
}

// GetEndpointName provides a mock function with no fields
func (_m *DownloadEndpoint) GetEndpointName() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetEndpointName")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetEndpointsList provides a mock function with no fields
func (_m *DownloadEndpoint) GetEndpointsList() map[common.MethodPath][]gin.HandlerFunc {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetEndpointsList")
	}

	var r0 map[common.MethodPath][]gin.HandlerFunc
	if rf, ok := ret.Get(0).(func() map[common.MethodPath][]gin.HandlerFunc); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[common.MethodPath][]gin.HandlerFunc)
		}
	}

	return r0
}

// GetGroupUrl provides a mock function with no fields
func (_m *DownloadEndpoint) GetGroupUrl() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetGroupUrl")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetPermissionsManager provides a mock function with no fields
func (_m *DownloadEndpoint) GetPermissionsManager() common.PermissionsManager {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetPermissionsManager")
	}

	var r0 common.PermissionsManager
	if rf, ok := ret.Get(0).(func() common.PermissionsManager); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(common.PermissionsManager)
		}
	}

	return r0
}

// InitDownload provides a mock function with given fields: c
func (_m *DownloadEndpoint) InitDownload(c *gin.Context) {
	_m.Called(c)
}

// NewDownloadEndpoint creates a new instance of DownloadEndpoint. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewDownloadEndpoint(t interface {
	mock.TestingT
	Cleanup(func())
}) *DownloadEndpoint {
	mock := &DownloadEndpoint{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
