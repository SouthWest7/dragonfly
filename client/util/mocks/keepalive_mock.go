// Code generated by MockGen. DO NOT EDIT.
// Source: keepalive.go
//
// Generated by this command:
//
//	mockgen -destination mocks/keepalive_mock.go -source keepalive.go -package mocks
//

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"
	time "time"

	gomock "go.uber.org/mock/gomock"
)

// MockKeepAlive is a mock of KeepAlive interface.
type MockKeepAlive struct {
	ctrl     *gomock.Controller
	recorder *MockKeepAliveMockRecorder
	isgomock struct{}
}

// MockKeepAliveMockRecorder is the mock recorder for MockKeepAlive.
type MockKeepAliveMockRecorder struct {
	mock *MockKeepAlive
}

// NewMockKeepAlive creates a new mock instance.
func NewMockKeepAlive(ctrl *gomock.Controller) *MockKeepAlive {
	mock := &MockKeepAlive{ctrl: ctrl}
	mock.recorder = &MockKeepAliveMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockKeepAlive) EXPECT() *MockKeepAliveMockRecorder {
	return m.recorder
}

// Alive mocks base method.
func (m *MockKeepAlive) Alive(alive time.Duration) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Alive", alive)
	ret0, _ := ret[0].(bool)
	return ret0
}

// Alive indicates an expected call of Alive.
func (mr *MockKeepAliveMockRecorder) Alive(alive any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Alive", reflect.TypeOf((*MockKeepAlive)(nil).Alive), alive)
}

// Keep mocks base method.
func (m *MockKeepAlive) Keep() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Keep")
}

// Keep indicates an expected call of Keep.
func (mr *MockKeepAliveMockRecorder) Keep() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Keep", reflect.TypeOf((*MockKeepAlive)(nil).Keep))
}
