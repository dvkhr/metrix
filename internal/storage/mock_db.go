// Code generated by MockGen. DO NOT EDIT.
// Source: internal/storage/dbstorage.go

// Package storage is a generated GoMock package.
package storage

import (
	context "context"
	sql "database/sql"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockDB is a mock of DB interface.
type MockDB struct {
	ctrl     *gomock.Controller
	recorder *MockDBMockRecorder
}

// MockDBMockRecorder is the mock recorder for MockDB.
type MockDBMockRecorder struct {
	mock *MockDB
}

// NewMockDB creates a new mock instance.
func NewMockDB(ctrl *gomock.Controller) *MockDB {
	mock := &MockDB{ctrl: ctrl}
	mock.recorder = &MockDBMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDB) EXPECT() *MockDBMockRecorder {
	return m.recorder
}

// BeginTx mocks base method.
func (m *MockDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BeginTx", ctx, opts)
	ret0, _ := ret[0].(*sql.Tx)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BeginTx indicates an expected call of BeginTx.
func (mr *MockDBMockRecorder) BeginTx(ctx, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BeginTx", reflect.TypeOf((*MockDB)(nil).BeginTx), ctx, opts)
}

// Close mocks base method.
func (m *MockDB) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockDBMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockDB)(nil).Close))
}

// Exec mocks base method.
func (m *MockDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{query}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Exec", varargs...)
	ret0, _ := ret[0].(sql.Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Exec indicates an expected call of Exec.
func (mr *MockDBMockRecorder) Exec(query interface{}, args ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{query}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Exec", reflect.TypeOf((*MockDB)(nil).Exec), varargs...)
}

// Ping mocks base method.
func (m *MockDB) Ping() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Ping")
	ret0, _ := ret[0].(error)
	return ret0
}

// Ping indicates an expected call of Ping.
func (mr *MockDBMockRecorder) Ping() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Ping", reflect.TypeOf((*MockDB)(nil).Ping))
}

// Prepare mocks base method.
func (m *MockDB) Prepare(query string) (*sql.Stmt, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Prepare", query)
	ret0, _ := ret[0].(*sql.Stmt)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Prepare indicates an expected call of Prepare.
func (mr *MockDBMockRecorder) Prepare(query interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Prepare", reflect.TypeOf((*MockDB)(nil).Prepare), query)
}

// MockStmt is a mock of Stmt interface.
type MockStmt struct {
	ctrl     *gomock.Controller
	recorder *MockStmtMockRecorder
}

// MockStmtMockRecorder is the mock recorder for MockStmt.
type MockStmtMockRecorder struct {
	mock *MockStmt
}

// NewMockStmt creates a new mock instance.
func NewMockStmt(ctrl *gomock.Controller) *MockStmt {
	mock := &MockStmt{ctrl: ctrl}
	mock.recorder = &MockStmtMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStmt) EXPECT() *MockStmtMockRecorder {
	return m.recorder
}

// Exec mocks base method.
func (m *MockStmt) Exec(args ...interface{}) (sql.Result, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Exec", varargs...)
	ret0, _ := ret[0].(sql.Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Exec indicates an expected call of Exec.
func (mr *MockStmtMockRecorder) Exec(args ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Exec", reflect.TypeOf((*MockStmt)(nil).Exec), args...)
}

// QueryRow mocks base method.
func (m *MockStmt) QueryRow(args ...interface{}) *sql.Row {
	m.ctrl.T.Helper()
	varargs := []interface{}{}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "QueryRow", varargs...)
	ret0, _ := ret[0].(*sql.Row)
	return ret0
}

// QueryRow indicates an expected call of QueryRow.
func (mr *MockStmtMockRecorder) QueryRow(args ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryRow", reflect.TypeOf((*MockStmt)(nil).QueryRow), args...)
}
