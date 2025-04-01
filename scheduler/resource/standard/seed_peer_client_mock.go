// Code generated by MockGen. DO NOT EDIT.
// Source: seed_peer_client.go
//
// Generated by this command:
//
//	mockgen -destination seed_peer_client_mock.go -source seed_peer_client.go -package standard
//

// Package standard is a generated GoMock package.
package standard

import (
	context "context"
	reflect "reflect"

	cdnsystem "d7y.io/api/v2/pkg/apis/cdnsystem/v1"
	common "d7y.io/api/v2/pkg/apis/common/v1"
	common0 "d7y.io/api/v2/pkg/apis/common/v2"
	dfdaemon "d7y.io/api/v2/pkg/apis/dfdaemon/v2"
	manager "d7y.io/api/v2/pkg/apis/manager/v2"
	config "d7y.io/dragonfly/v2/scheduler/config"
	gomock "go.uber.org/mock/gomock"
	grpc "google.golang.org/grpc"
)

// MockSeedPeerClient is a mock of SeedPeerClient interface.
type MockSeedPeerClient struct {
	ctrl     *gomock.Controller
	recorder *MockSeedPeerClientMockRecorder
	isgomock struct{}
}

// MockSeedPeerClientMockRecorder is the mock recorder for MockSeedPeerClient.
type MockSeedPeerClientMockRecorder struct {
	mock *MockSeedPeerClient
}

// NewMockSeedPeerClient creates a new mock instance.
func NewMockSeedPeerClient(ctrl *gomock.Controller) *MockSeedPeerClient {
	mock := &MockSeedPeerClient{ctrl: ctrl}
	mock.recorder = &MockSeedPeerClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSeedPeerClient) EXPECT() *MockSeedPeerClientMockRecorder {
	return m.recorder
}

// Addrs mocks base method.
func (m *MockSeedPeerClient) Addrs() []string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Addrs")
	ret0, _ := ret[0].([]string)
	return ret0
}

// Addrs indicates an expected call of Addrs.
func (mr *MockSeedPeerClientMockRecorder) Addrs() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Addrs", reflect.TypeOf((*MockSeedPeerClient)(nil).Addrs))
}

// Close mocks base method.
func (m *MockSeedPeerClient) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockSeedPeerClientMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockSeedPeerClient)(nil).Close))
}

// DeletePersistentCacheTask mocks base method.
func (m *MockSeedPeerClient) DeletePersistentCacheTask(arg0 context.Context, arg1 *dfdaemon.DeletePersistentCacheTaskRequest, arg2 ...grpc.CallOption) error {
	m.ctrl.T.Helper()
	varargs := []any{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "DeletePersistentCacheTask", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeletePersistentCacheTask indicates an expected call of DeletePersistentCacheTask.
func (mr *MockSeedPeerClientMockRecorder) DeletePersistentCacheTask(arg0, arg1 any, arg2 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeletePersistentCacheTask", reflect.TypeOf((*MockSeedPeerClient)(nil).DeletePersistentCacheTask), varargs...)
}

// DeleteTask mocks base method.
func (m *MockSeedPeerClient) DeleteTask(arg0 context.Context, arg1 *dfdaemon.DeleteTaskRequest, arg2 ...grpc.CallOption) error {
	m.ctrl.T.Helper()
	varargs := []any{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "DeleteTask", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteTask indicates an expected call of DeleteTask.
func (mr *MockSeedPeerClientMockRecorder) DeleteTask(arg0, arg1 any, arg2 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteTask", reflect.TypeOf((*MockSeedPeerClient)(nil).DeleteTask), varargs...)
}

// DownloadPersistentCacheTask mocks base method.
func (m *MockSeedPeerClient) DownloadPersistentCacheTask(arg0 context.Context, arg1 *dfdaemon.DownloadPersistentCacheTaskRequest, arg2 ...grpc.CallOption) (dfdaemon.DfdaemonUpload_DownloadPersistentCacheTaskClient, error) {
	m.ctrl.T.Helper()
	varargs := []any{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "DownloadPersistentCacheTask", varargs...)
	ret0, _ := ret[0].(dfdaemon.DfdaemonUpload_DownloadPersistentCacheTaskClient)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DownloadPersistentCacheTask indicates an expected call of DownloadPersistentCacheTask.
func (mr *MockSeedPeerClientMockRecorder) DownloadPersistentCacheTask(arg0, arg1 any, arg2 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DownloadPersistentCacheTask", reflect.TypeOf((*MockSeedPeerClient)(nil).DownloadPersistentCacheTask), varargs...)
}

// DownloadPiece mocks base method.
func (m *MockSeedPeerClient) DownloadPiece(arg0 context.Context, arg1 *dfdaemon.DownloadPieceRequest, arg2 ...grpc.CallOption) (*dfdaemon.DownloadPieceResponse, error) {
	m.ctrl.T.Helper()
	varargs := []any{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "DownloadPiece", varargs...)
	ret0, _ := ret[0].(*dfdaemon.DownloadPieceResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DownloadPiece indicates an expected call of DownloadPiece.
func (mr *MockSeedPeerClientMockRecorder) DownloadPiece(arg0, arg1 any, arg2 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DownloadPiece", reflect.TypeOf((*MockSeedPeerClient)(nil).DownloadPiece), varargs...)
}

// DownloadTask mocks base method.
func (m *MockSeedPeerClient) DownloadTask(arg0 context.Context, arg1 string, arg2 *dfdaemon.DownloadTaskRequest, arg3 ...grpc.CallOption) (dfdaemon.DfdaemonUpload_DownloadTaskClient, error) {
	m.ctrl.T.Helper()
	varargs := []any{arg0, arg1, arg2}
	for _, a := range arg3 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "DownloadTask", varargs...)
	ret0, _ := ret[0].(dfdaemon.DfdaemonUpload_DownloadTaskClient)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DownloadTask indicates an expected call of DownloadTask.
func (mr *MockSeedPeerClientMockRecorder) DownloadTask(arg0, arg1, arg2 any, arg3 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0, arg1, arg2}, arg3...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DownloadTask", reflect.TypeOf((*MockSeedPeerClient)(nil).DownloadTask), varargs...)
}

// GetPieceTasks mocks base method.
func (m *MockSeedPeerClient) GetPieceTasks(arg0 context.Context, arg1 *common.PieceTaskRequest, arg2 ...grpc.CallOption) (*common.PiecePacket, error) {
	m.ctrl.T.Helper()
	varargs := []any{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetPieceTasks", varargs...)
	ret0, _ := ret[0].(*common.PiecePacket)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPieceTasks indicates an expected call of GetPieceTasks.
func (mr *MockSeedPeerClientMockRecorder) GetPieceTasks(arg0, arg1 any, arg2 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPieceTasks", reflect.TypeOf((*MockSeedPeerClient)(nil).GetPieceTasks), varargs...)
}

// ObtainSeeds mocks base method.
func (m *MockSeedPeerClient) ObtainSeeds(arg0 context.Context, arg1 *cdnsystem.SeedRequest, arg2 ...grpc.CallOption) (cdnsystem.Seeder_ObtainSeedsClient, error) {
	m.ctrl.T.Helper()
	varargs := []any{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ObtainSeeds", varargs...)
	ret0, _ := ret[0].(cdnsystem.Seeder_ObtainSeedsClient)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ObtainSeeds indicates an expected call of ObtainSeeds.
func (mr *MockSeedPeerClientMockRecorder) ObtainSeeds(arg0, arg1 any, arg2 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ObtainSeeds", reflect.TypeOf((*MockSeedPeerClient)(nil).ObtainSeeds), varargs...)
}

// OnNotify mocks base method.
func (m *MockSeedPeerClient) OnNotify(arg0 *config.DynconfigData) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "OnNotify", arg0)
}

// OnNotify indicates an expected call of OnNotify.
func (mr *MockSeedPeerClientMockRecorder) OnNotify(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OnNotify", reflect.TypeOf((*MockSeedPeerClient)(nil).OnNotify), arg0)
}

// SeedPeers mocks base method.
func (m *MockSeedPeerClient) SeedPeers() []*manager.SeedPeer {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SeedPeers")
	ret0, _ := ret[0].([]*manager.SeedPeer)
	return ret0
}

// SeedPeers indicates an expected call of SeedPeers.
func (mr *MockSeedPeerClientMockRecorder) SeedPeers() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SeedPeers", reflect.TypeOf((*MockSeedPeerClient)(nil).SeedPeers))
}

// StatPersistentCacheTask mocks base method.
func (m *MockSeedPeerClient) StatPersistentCacheTask(arg0 context.Context, arg1 *dfdaemon.StatPersistentCacheTaskRequest, arg2 ...grpc.CallOption) (*common0.PersistentCacheTask, error) {
	m.ctrl.T.Helper()
	varargs := []any{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "StatPersistentCacheTask", varargs...)
	ret0, _ := ret[0].(*common0.PersistentCacheTask)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// StatPersistentCacheTask indicates an expected call of StatPersistentCacheTask.
func (mr *MockSeedPeerClientMockRecorder) StatPersistentCacheTask(arg0, arg1 any, arg2 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StatPersistentCacheTask", reflect.TypeOf((*MockSeedPeerClient)(nil).StatPersistentCacheTask), varargs...)
}

// StatTask mocks base method.
func (m *MockSeedPeerClient) StatTask(arg0 context.Context, arg1 *dfdaemon.StatTaskRequest, arg2 ...grpc.CallOption) (*common0.Task, error) {
	m.ctrl.T.Helper()
	varargs := []any{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "StatTask", varargs...)
	ret0, _ := ret[0].(*common0.Task)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// StatTask indicates an expected call of StatTask.
func (mr *MockSeedPeerClientMockRecorder) StatTask(arg0, arg1 any, arg2 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StatTask", reflect.TypeOf((*MockSeedPeerClient)(nil).StatTask), varargs...)
}

// SyncPieceTasks mocks base method.
func (m *MockSeedPeerClient) SyncPieceTasks(arg0 context.Context, arg1 *common.PieceTaskRequest, arg2 ...grpc.CallOption) (cdnsystem.Seeder_SyncPieceTasksClient, error) {
	m.ctrl.T.Helper()
	varargs := []any{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "SyncPieceTasks", varargs...)
	ret0, _ := ret[0].(cdnsystem.Seeder_SyncPieceTasksClient)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SyncPieceTasks indicates an expected call of SyncPieceTasks.
func (mr *MockSeedPeerClientMockRecorder) SyncPieceTasks(arg0, arg1 any, arg2 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SyncPieceTasks", reflect.TypeOf((*MockSeedPeerClient)(nil).SyncPieceTasks), varargs...)
}

// SyncPieces mocks base method.
func (m *MockSeedPeerClient) SyncPieces(arg0 context.Context, arg1 *dfdaemon.SyncPiecesRequest, arg2 ...grpc.CallOption) (dfdaemon.DfdaemonUpload_SyncPiecesClient, error) {
	m.ctrl.T.Helper()
	varargs := []any{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "SyncPieces", varargs...)
	ret0, _ := ret[0].(dfdaemon.DfdaemonUpload_SyncPiecesClient)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SyncPieces indicates an expected call of SyncPieces.
func (mr *MockSeedPeerClientMockRecorder) SyncPieces(arg0, arg1 any, arg2 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SyncPieces", reflect.TypeOf((*MockSeedPeerClient)(nil).SyncPieces), varargs...)
}

// UpdatePersistentCacheTask mocks base method.
func (m *MockSeedPeerClient) UpdatePersistentCacheTask(arg0 context.Context, arg1 *dfdaemon.UpdatePersistentCacheTaskRequest, arg2 ...grpc.CallOption) error {
	m.ctrl.T.Helper()
	varargs := []any{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "UpdatePersistentCacheTask", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdatePersistentCacheTask indicates an expected call of UpdatePersistentCacheTask.
func (mr *MockSeedPeerClientMockRecorder) UpdatePersistentCacheTask(arg0, arg1 any, arg2 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdatePersistentCacheTask", reflect.TypeOf((*MockSeedPeerClient)(nil).UpdatePersistentCacheTask), varargs...)
}
