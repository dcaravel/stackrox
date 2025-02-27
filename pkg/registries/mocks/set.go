// Code generated by MockGen. DO NOT EDIT.
// Source: set.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	storage "github.com/stackrox/rox/generated/storage"
	types "github.com/stackrox/rox/pkg/registries/types"
)

// MockSet is a mock of Set interface.
type MockSet struct {
	ctrl     *gomock.Controller
	recorder *MockSetMockRecorder
}

// MockSetMockRecorder is the mock recorder for MockSet.
type MockSetMockRecorder struct {
	mock *MockSet
}

// NewMockSet creates a new mock instance.
func NewMockSet(ctrl *gomock.Controller) *MockSet {
	mock := &MockSet{ctrl: ctrl}
	mock.recorder = &MockSetMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSet) EXPECT() *MockSetMockRecorder {
	return m.recorder
}

// Clear mocks base method.
func (m *MockSet) Clear() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Clear")
}

// Clear indicates an expected call of Clear.
func (mr *MockSetMockRecorder) Clear() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Clear", reflect.TypeOf((*MockSet)(nil).Clear))
}

// GetAll mocks base method.
func (m *MockSet) GetAll() []types.ImageRegistry {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAll")
	ret0, _ := ret[0].([]types.ImageRegistry)
	return ret0
}

// GetAll indicates an expected call of GetAll.
func (mr *MockSetMockRecorder) GetAll() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAll", reflect.TypeOf((*MockSet)(nil).GetAll))
}

// GetRegistryByImage mocks base method.
func (m *MockSet) GetRegistryByImage(image *storage.Image) types.Registry {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRegistryByImage", image)
	ret0, _ := ret[0].(types.Registry)
	return ret0
}

// GetRegistryByImage indicates an expected call of GetRegistryByImage.
func (mr *MockSetMockRecorder) GetRegistryByImage(image interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRegistryByImage", reflect.TypeOf((*MockSet)(nil).GetRegistryByImage), image)
}

// GetRegistryMetadataByImage mocks base method.
func (m *MockSet) GetRegistryMetadataByImage(image *storage.Image) *types.Config {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRegistryMetadataByImage", image)
	ret0, _ := ret[0].(*types.Config)
	return ret0
}

// GetRegistryMetadataByImage indicates an expected call of GetRegistryMetadataByImage.
func (mr *MockSetMockRecorder) GetRegistryMetadataByImage(image interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRegistryMetadataByImage", reflect.TypeOf((*MockSet)(nil).GetRegistryMetadataByImage), image)
}

// IsEmpty mocks base method.
func (m *MockSet) IsEmpty() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsEmpty")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsEmpty indicates an expected call of IsEmpty.
func (mr *MockSetMockRecorder) IsEmpty() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsEmpty", reflect.TypeOf((*MockSet)(nil).IsEmpty))
}

// Match mocks base method.
func (m *MockSet) Match(image *storage.ImageName) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Match", image)
	ret0, _ := ret[0].(bool)
	return ret0
}

// Match indicates an expected call of Match.
func (mr *MockSetMockRecorder) Match(image interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Match", reflect.TypeOf((*MockSet)(nil).Match), image)
}

// RemoveImageIntegration mocks base method.
func (m *MockSet) RemoveImageIntegration(id string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveImageIntegration", id)
	ret0, _ := ret[0].(error)
	return ret0
}

// RemoveImageIntegration indicates an expected call of RemoveImageIntegration.
func (mr *MockSetMockRecorder) RemoveImageIntegration(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveImageIntegration", reflect.TypeOf((*MockSet)(nil).RemoveImageIntegration), id)
}

// UpdateImageIntegration mocks base method.
func (m *MockSet) UpdateImageIntegration(integration *storage.ImageIntegration) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateImageIntegration", integration)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateImageIntegration indicates an expected call of UpdateImageIntegration.
func (mr *MockSetMockRecorder) UpdateImageIntegration(integration interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateImageIntegration", reflect.TypeOf((*MockSet)(nil).UpdateImageIntegration), integration)
}
