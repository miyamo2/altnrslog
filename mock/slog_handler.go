// Code generated by MockGen. DO NOT EDIT.
// Source: log/slog (interfaces: Handler)
//
// Generated by this command:
//
//	mockgen -destination mock/slog_handler.go log/slog Handler
//

// Package mock_slog is a generated GoMock package.
package mock_slog

import (
	context "context"
	slog "log/slog"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockHandler is a mock of Handler interface.
type MockHandler struct {
	ctrl     *gomock.Controller
	recorder *MockHandlerMockRecorder
}

// MockHandlerMockRecorder is the mock recorder for MockHandler.
type MockHandlerMockRecorder struct {
	mock *MockHandler
}

// NewMockHandler creates a new mock instance.
func NewMockHandler(ctrl *gomock.Controller) *MockHandler {
	mock := &MockHandler{ctrl: ctrl}
	mock.recorder = &MockHandlerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockHandler) EXPECT() *MockHandlerMockRecorder {
	return m.recorder
}

// Enabled mocks base method.
func (m *MockHandler) Enabled(arg0 context.Context, arg1 slog.Level) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Enabled", arg0, arg1)
	ret0, _ := ret[0].(bool)
	return ret0
}

// Enabled indicates an expected call of Enabled.
func (mr *MockHandlerMockRecorder) Enabled(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Enabled", reflect.TypeOf((*MockHandler)(nil).Enabled), arg0, arg1)
}

// Handle mocks base method.
func (m *MockHandler) Handle(arg0 context.Context, arg1 slog.Record) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Handle", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Handle indicates an expected call of Handle.
func (mr *MockHandlerMockRecorder) Handle(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Handle", reflect.TypeOf((*MockHandler)(nil).Handle), arg0, arg1)
}

// WithAttrs mocks base method.
func (m *MockHandler) WithAttrs(arg0 []slog.Attr) slog.Handler {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WithAttrs", arg0)
	ret0, _ := ret[0].(slog.Handler)
	return ret0
}

// WithAttrs indicates an expected call of WithAttrs.
func (mr *MockHandlerMockRecorder) WithAttrs(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WithAttrs", reflect.TypeOf((*MockHandler)(nil).WithAttrs), arg0)
}

// WithGroup mocks base method.
func (m *MockHandler) WithGroup(arg0 string) slog.Handler {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WithGroup", arg0)
	ret0, _ := ret[0].(slog.Handler)
	return ret0
}

// WithGroup indicates an expected call of WithGroup.
func (mr *MockHandlerMockRecorder) WithGroup(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WithGroup", reflect.TypeOf((*MockHandler)(nil).WithGroup), arg0)
}
