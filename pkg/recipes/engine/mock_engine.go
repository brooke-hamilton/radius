// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/radius-project/radius/pkg/recipes/engine (interfaces: Engine)
//
// Generated by this command:
//
//	mockgen -destination=./mock_engine.go -package=engine -self_package github.com/radius-project/radius/pkg/recipes/engine github.com/radius-project/radius/pkg/recipes/engine Engine
//

// Package engine is a generated GoMock package.
package engine

import (
	context "context"
	reflect "reflect"

	recipes "github.com/radius-project/radius/pkg/recipes"
	gomock "go.uber.org/mock/gomock"
)

// MockEngine is a mock of Engine interface.
type MockEngine struct {
	ctrl     *gomock.Controller
	recorder *MockEngineMockRecorder
}

// MockEngineMockRecorder is the mock recorder for MockEngine.
type MockEngineMockRecorder struct {
	mock *MockEngine
}

// NewMockEngine creates a new mock instance.
func NewMockEngine(ctrl *gomock.Controller) *MockEngine {
	mock := &MockEngine{ctrl: ctrl}
	mock.recorder = &MockEngineMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockEngine) EXPECT() *MockEngineMockRecorder {
	return m.recorder
}

// Delete mocks base method.
func (m *MockEngine) Delete(arg0 context.Context, arg1 DeleteOptions) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockEngineMockRecorder) Delete(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockEngine)(nil).Delete), arg0, arg1)
}

// Execute mocks base method.
func (m *MockEngine) Execute(arg0 context.Context, arg1 ExecuteOptions) (*recipes.RecipeOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Execute", arg0, arg1)
	ret0, _ := ret[0].(*recipes.RecipeOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Execute indicates an expected call of Execute.
func (mr *MockEngineMockRecorder) Execute(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Execute", reflect.TypeOf((*MockEngine)(nil).Execute), arg0, arg1)
}

// GetRecipeMetadata mocks base method.
func (m *MockEngine) GetRecipeMetadata(arg0 context.Context, arg1 recipes.EnvironmentDefinition) (map[string]any, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRecipeMetadata", arg0, arg1)
	ret0, _ := ret[0].(map[string]any)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRecipeMetadata indicates an expected call of GetRecipeMetadata.
func (mr *MockEngineMockRecorder) GetRecipeMetadata(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRecipeMetadata", reflect.TypeOf((*MockEngine)(nil).GetRecipeMetadata), arg0, arg1)
}
