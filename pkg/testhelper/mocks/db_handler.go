// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	models "content-service-api/models"

	primitive "go.mongodb.org/mongo-driver/bson/primitive"
)

// DBHandler is an autogenerated mock type for the DBHandler type
type DBHandler struct {
	mock.Mock
}

// DeleteFile provides a mock function with given fields: ctx, fileID
func (_m *DBHandler) DeleteFile(ctx context.Context, fileID primitive.ObjectID) error {
	ret := _m.Called(ctx, fileID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, primitive.ObjectID) error); ok {
		r0 = rf(ctx, fileID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetFile provides a mock function with given fields: ctx, fileID
func (_m *DBHandler) GetFile(ctx context.Context, fileID primitive.ObjectID) ([]byte, error) {
	ret := _m.Called(ctx, fileID)

	var r0 []byte
	if rf, ok := ret.Get(0).(func(context.Context, primitive.ObjectID) []byte); ok {
		r0 = rf(ctx, fileID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, primitive.ObjectID) error); ok {
		r1 = rf(ctx, fileID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetFiles provides a mock function with given fields: ctx, query
func (_m *DBHandler) GetFiles(ctx context.Context, query map[string]interface{}) ([]models.FileResponse, error) {
	ret := _m.Called(ctx, query)

	var r0 []models.FileResponse
	if rf, ok := ret.Get(0).(func(context.Context, map[string]interface{}) []models.FileResponse); ok {
		r0 = rf(ctx, query)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]models.FileResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, map[string]interface{}) error); ok {
		r1 = rf(ctx, query)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Ping provides a mock function with given fields: ctx
func (_m *DBHandler) Ping(ctx context.Context) error {
	ret := _m.Called(ctx)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateFileInfo provides a mock function with given fields: ctx, fileID, updateRequest
func (_m *DBHandler) UpdateFileInfo(ctx context.Context, fileID primitive.ObjectID, updateRequest map[string]interface{}) error {
	ret := _m.Called(ctx, fileID, updateRequest)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, primitive.ObjectID, map[string]interface{}) error); ok {
		r0 = rf(ctx, fileID, updateRequest)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UploadFile provides a mock function with given fields: ctx, uploadRequest, fileBytes
func (_m *DBHandler) UploadFile(ctx context.Context, uploadRequest *models.FileRequest, fileBytes []byte) error {
	ret := _m.Called(ctx, uploadRequest, fileBytes)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *models.FileRequest, []byte) error); ok {
		r0 = rf(ctx, uploadRequest, fileBytes)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
