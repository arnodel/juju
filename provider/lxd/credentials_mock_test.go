// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/juju/juju/provider/lxd (interfaces: CertificateReadWriter,CertificateGenerator,NetLookup)

// Package lxd is a generated GoMock package.
package lxd

import (
	net "net"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockCertificateReadWriter is a mock of CertificateReadWriter interface
type MockCertificateReadWriter struct {
	ctrl     *gomock.Controller
	recorder *MockCertificateReadWriterMockRecorder
}

// MockCertificateReadWriterMockRecorder is the mock recorder for MockCertificateReadWriter
type MockCertificateReadWriterMockRecorder struct {
	mock *MockCertificateReadWriter
}

// NewMockCertificateReadWriter creates a new mock instance
func NewMockCertificateReadWriter(ctrl *gomock.Controller) *MockCertificateReadWriter {
	mock := &MockCertificateReadWriter{ctrl: ctrl}
	mock.recorder = &MockCertificateReadWriterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockCertificateReadWriter) EXPECT() *MockCertificateReadWriterMockRecorder {
	return m.recorder
}

// Read mocks base method
func (m *MockCertificateReadWriter) Read(arg0 string) ([]byte, []byte, error) {
	ret := m.ctrl.Call(m, "Read", arg0)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].([]byte)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Read indicates an expected call of Read
func (mr *MockCertificateReadWriterMockRecorder) Read(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Read", reflect.TypeOf((*MockCertificateReadWriter)(nil).Read), arg0)
}

// Write mocks base method
func (m *MockCertificateReadWriter) Write(arg0 string, arg1, arg2 []byte) error {
	ret := m.ctrl.Call(m, "Write", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// Write indicates an expected call of Write
func (mr *MockCertificateReadWriterMockRecorder) Write(arg0, arg1, arg2 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Write", reflect.TypeOf((*MockCertificateReadWriter)(nil).Write), arg0, arg1, arg2)
}

// MockCertificateGenerator is a mock of CertificateGenerator interface
type MockCertificateGenerator struct {
	ctrl     *gomock.Controller
	recorder *MockCertificateGeneratorMockRecorder
}

// MockCertificateGeneratorMockRecorder is the mock recorder for MockCertificateGenerator
type MockCertificateGeneratorMockRecorder struct {
	mock *MockCertificateGenerator
}

// NewMockCertificateGenerator creates a new mock instance
func NewMockCertificateGenerator(ctrl *gomock.Controller) *MockCertificateGenerator {
	mock := &MockCertificateGenerator{ctrl: ctrl}
	mock.recorder = &MockCertificateGeneratorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockCertificateGenerator) EXPECT() *MockCertificateGeneratorMockRecorder {
	return m.recorder
}

// Generate mocks base method
func (m *MockCertificateGenerator) Generate(arg0 bool) ([]byte, []byte, error) {
	ret := m.ctrl.Call(m, "Generate", arg0)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].([]byte)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Generate indicates an expected call of Generate
func (mr *MockCertificateGeneratorMockRecorder) Generate(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Generate", reflect.TypeOf((*MockCertificateGenerator)(nil).Generate), arg0)
}

// MockNetLookup is a mock of NetLookup interface
type MockNetLookup struct {
	ctrl     *gomock.Controller
	recorder *MockNetLookupMockRecorder
}

// MockNetLookupMockRecorder is the mock recorder for MockNetLookup
type MockNetLookupMockRecorder struct {
	mock *MockNetLookup
}

// NewMockNetLookup creates a new mock instance
func NewMockNetLookup(ctrl *gomock.Controller) *MockNetLookup {
	mock := &MockNetLookup{ctrl: ctrl}
	mock.recorder = &MockNetLookupMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockNetLookup) EXPECT() *MockNetLookupMockRecorder {
	return m.recorder
}

// InterfaceAddrs mocks base method
func (m *MockNetLookup) InterfaceAddrs() ([]net.Addr, error) {
	ret := m.ctrl.Call(m, "InterfaceAddrs")
	ret0, _ := ret[0].([]net.Addr)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// InterfaceAddrs indicates an expected call of InterfaceAddrs
func (mr *MockNetLookupMockRecorder) InterfaceAddrs() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InterfaceAddrs", reflect.TypeOf((*MockNetLookup)(nil).InterfaceAddrs))
}

// LookupHost mocks base method
func (m *MockNetLookup) LookupHost(arg0 string) ([]string, error) {
	ret := m.ctrl.Call(m, "LookupHost", arg0)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LookupHost indicates an expected call of LookupHost
func (mr *MockNetLookupMockRecorder) LookupHost(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LookupHost", reflect.TypeOf((*MockNetLookup)(nil).LookupHost), arg0)
}
