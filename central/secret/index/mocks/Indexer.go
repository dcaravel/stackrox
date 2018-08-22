// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"
import v1 "github.com/stackrox/rox/generated/api/v1"

// Indexer is an autogenerated mock type for the Indexer type
type Indexer struct {
	mock.Mock
}

// SecretAndRelationship provides a mock function with given fields: sar
func (_m *Indexer) SecretAndRelationship(sar *v1.SecretAndRelationship) error {
	ret := _m.Called(sar)

	var r0 error
	if rf, ok := ret.Get(0).(func(*v1.SecretAndRelationship) error); ok {
		r0 = rf(sar)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
