// Code generated by MockGen. DO NOT EDIT.
// Source: seed_peer.go
//
// Generated by this command:
//
//	mockgen -destination seed_peer_mock.go -source seed_peer.go -package standard
//

// Package standard is a generated GoMock package.
package standard

import (
	context "context"
	reflect "reflect"

	dfdaemon "d7y.io/api/v2/pkg/apis/dfdaemon/v2"
	scheduler "d7y.io/api/v2/pkg/apis/scheduler/v1"
	http "d7y.io/dragonfly/v2/pkg/net/http"
	gomock "go.uber.org/mock/gomock"
)

// MockSeedPeer is a mock of SeedPeer interface.
type MockSeedPeer struct {
	ctrl     *gomock.Controller
	recorder *MockSeedPeerMockRecorder
	isgomock struct{}
}

// MockSeedPeerMockRecorder is the mock recorder for MockSeedPeer.
type MockSeedPeerMockRecorder struct {
	mock *MockSeedPeer
}

// NewMockSeedPeer creates a new mock instance.
func NewMockSeedPeer(ctrl *gomock.Controller) *MockSeedPeer {
	mock := &MockSeedPeer{ctrl: ctrl}
	mock.recorder = &MockSeedPeerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSeedPeer) EXPECT() *MockSeedPeerMockRecorder {
	return m.recorder
}

// Client mocks base method.
func (m *MockSeedPeer) Client() SeedPeerClient {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Client")
	ret0, _ := ret[0].(SeedPeerClient)
	return ret0
}

// Client indicates an expected call of Client.
func (mr *MockSeedPeerMockRecorder) Client() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Client", reflect.TypeOf((*MockSeedPeer)(nil).Client))
}

// Stop mocks base method.
func (m *MockSeedPeer) Stop() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Stop")
	ret0, _ := ret[0].(error)
	return ret0
}

// Stop indicates an expected call of Stop.
func (mr *MockSeedPeerMockRecorder) Stop() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockSeedPeer)(nil).Stop))
}

// TriggerDownloadTask mocks base method.
func (m *MockSeedPeer) TriggerDownloadTask(arg0 context.Context, arg1 string, arg2 *dfdaemon.DownloadTaskRequest) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TriggerDownloadTask", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// TriggerDownloadTask indicates an expected call of TriggerDownloadTask.
func (mr *MockSeedPeerMockRecorder) TriggerDownloadTask(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TriggerDownloadTask", reflect.TypeOf((*MockSeedPeer)(nil).TriggerDownloadTask), arg0, arg1, arg2)
}

// TriggerTask mocks base method.
func (m *MockSeedPeer) TriggerTask(arg0 context.Context, arg1 *http.Range, arg2 *Task) (*Peer, *scheduler.PeerResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TriggerTask", arg0, arg1, arg2)
	ret0, _ := ret[0].(*Peer)
	ret1, _ := ret[1].(*scheduler.PeerResult)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// TriggerTask indicates an expected call of TriggerTask.
func (mr *MockSeedPeerMockRecorder) TriggerTask(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TriggerTask", reflect.TypeOf((*MockSeedPeer)(nil).TriggerTask), arg0, arg1, arg2)
}
