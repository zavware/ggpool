// Code generated by MockGen. DO NOT EDIT.
// Source: clock.go

// Package mocks is a generated GoMock package.
package mocks

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
	time "time"
)

// Clock is a mock of clock interface
type Clock struct {
	ctrl     *gomock.Controller
	recorder *ClockMockRecorder
}

// ClockMockRecorder is the mock recorder for Clock
type ClockMockRecorder struct {
	mock *Clock
}

// NewClock creates a new mock instance
func NewClock(ctrl *gomock.Controller) *Clock {
	mock := &Clock{ctrl: ctrl}
	mock.recorder = &ClockMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *Clock) EXPECT() *ClockMockRecorder {
	return m.recorder
}

// Now mocks base method
func (m *Clock) Now() time.Time {
	ret := m.ctrl.Call(m, "Now")
	ret0, _ := ret[0].(time.Time)
	return ret0
}

// Now indicates an expected call of Now
func (mr *ClockMockRecorder) Now() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Now", reflect.TypeOf((*Clock)(nil).Now))
}
