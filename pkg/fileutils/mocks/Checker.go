// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"
import os "os"

// Checker is an autogenerated mock type for the Checker type
type Checker struct {
	mock.Mock
}

// Exists provides a mock function with given fields: path
func (_m *Checker) Exists(path string) bool {
	ret := _m.Called(path)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(path)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Filemode provides a mock function with given fields: path
func (_m *Checker) Filemode(path string) (os.FileMode, error) {
	ret := _m.Called(path)

	var r0 os.FileMode
	if rf, ok := ret.Get(0).(func(string) os.FileMode); ok {
		r0 = rf(path)
	} else {
		r0 = ret.Get(0).(os.FileMode)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(path)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}