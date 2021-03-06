// Code generated by MockGen. DO NOT EDIT.
// Source: service.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	gomock "github.com/golang/mock/gomock"
)

// MockTGBotAPI is a mock of TGBotAPI interface.
type MockTGBotAPI struct {
	ctrl     *gomock.Controller
	recorder *MockTGBotAPIMockRecorder
}

// MockTGBotAPIMockRecorder is the mock recorder for MockTGBotAPI.
type MockTGBotAPIMockRecorder struct {
	mock *MockTGBotAPI
}

// NewMockTGBotAPI creates a new mock instance.
func NewMockTGBotAPI(ctrl *gomock.Controller) *MockTGBotAPI {
	mock := &MockTGBotAPI{ctrl: ctrl}
	mock.recorder = &MockTGBotAPIMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTGBotAPI) EXPECT() *MockTGBotAPIMockRecorder {
	return m.recorder
}

// Request mocks base method.
func (m *MockTGBotAPI) Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Request", c)
	ret0, _ := ret[0].(*tgbotapi.APIResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Request indicates an expected call of Request.
func (mr *MockTGBotAPIMockRecorder) Request(c interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Request", reflect.TypeOf((*MockTGBotAPI)(nil).Request), c)
}

// Send mocks base method.
func (m *MockTGBotAPI) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Send", c)
	ret0, _ := ret[0].(tgbotapi.Message)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Send indicates an expected call of Send.
func (mr *MockTGBotAPIMockRecorder) Send(c interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockTGBotAPI)(nil).Send), c)
}

// StopReceivingUpdates mocks base method.
func (m *MockTGBotAPI) StopReceivingUpdates() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "StopReceivingUpdates")
}

// StopReceivingUpdates indicates an expected call of StopReceivingUpdates.
func (mr *MockTGBotAPIMockRecorder) StopReceivingUpdates() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StopReceivingUpdates", reflect.TypeOf((*MockTGBotAPI)(nil).StopReceivingUpdates))
}
