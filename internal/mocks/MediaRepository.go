// Code generated by mockery v2.53.3. DO NOT EDIT.

package mocks

import (
	model "data-storage-svc/internal/model"

	mock "github.com/stretchr/testify/mock"

	primitive "go.mongodb.org/mongo-driver/bson/primitive"
)

// MediaRepository is an autogenerated mock type for the MediaRepository type
type MediaRepository struct {
	mock.Mock
}

// Create provides a mock function with given fields: media
func (_m *MediaRepository) Create(media *model.Media) (*primitive.ObjectID, error) {
	ret := _m.Called(media)

	if len(ret) == 0 {
		panic("no return value specified for Create")
	}

	var r0 *primitive.ObjectID
	var r1 error
	if rf, ok := ret.Get(0).(func(*model.Media) (*primitive.ObjectID, error)); ok {
		return rf(media)
	}
	if rf, ok := ret.Get(0).(func(*model.Media) *primitive.ObjectID); ok {
		r0 = rf(media)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*primitive.ObjectID)
		}
	}

	if rf, ok := ret.Get(1).(func(*model.Media) error); ok {
		r1 = rf(media)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Delete provides a mock function with given fields: mediaId
func (_m *MediaRepository) Delete(mediaId *primitive.ObjectID) error {
	ret := _m.Called(mediaId)

	if len(ret) == 0 {
		panic("no return value specified for Delete")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(*primitive.ObjectID) error); ok {
		r0 = rf(mediaId)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Get provides a mock function with given fields: mediaId
func (_m *MediaRepository) Get(mediaId *primitive.ObjectID) (*model.Media, error) {
	ret := _m.Called(mediaId)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 *model.Media
	var r1 error
	if rf, ok := ret.Get(0).(func(*primitive.ObjectID) (*model.Media, error)); ok {
		return rf(mediaId)
	}
	if rf, ok := ret.Get(0).(func(*primitive.ObjectID) *model.Media); ok {
		r0 = rf(mediaId)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Media)
		}
	}

	if rf, ok := ret.Get(1).(func(*primitive.ObjectID) error); ok {
		r1 = rf(mediaId)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetAllUploadedBy provides a mock function with given fields: userId
func (_m *MediaRepository) GetAllUploadedBy(userId *primitive.ObjectID) ([]model.Media, error) {
	ret := _m.Called(userId)

	if len(ret) == 0 {
		panic("no return value specified for GetAllUploadedBy")
	}

	var r0 []model.Media
	var r1 error
	if rf, ok := ret.Get(0).(func(*primitive.ObjectID) ([]model.Media, error)); ok {
		return rf(userId)
	}
	if rf, ok := ret.Get(0).(func(*primitive.ObjectID) []model.Media); ok {
		r0 = rf(userId)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]model.Media)
		}
	}

	if rf, ok := ret.Get(1).(func(*primitive.ObjectID) error); ok {
		r1 = rf(userId)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Update provides a mock function with given fields: mediaId, update
func (_m *MediaRepository) Update(mediaId *primitive.ObjectID, update primitive.M) error {
	ret := _m.Called(mediaId, update)

	if len(ret) == 0 {
		panic("no return value specified for Update")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(*primitive.ObjectID, primitive.M) error); ok {
		r0 = rf(mediaId, update)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewMediaRepository creates a new instance of MediaRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMediaRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MediaRepository {
	mock := &MediaRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
