// Code generated by MockGen. DO NOT EDIT.
// Source: peertask_manager.go
//
// Generated by this command:
//
//	mockgen -destination peertask_manager_mock.go -source peertask_manager.go -package peer
//

// Package peer is a generated GoMock package.
package peer

import (
	context "context"
	io "io"
	reflect "reflect"

	common "d7y.io/api/v2/pkg/apis/common/v1"
	scheduler "d7y.io/api/v2/pkg/apis/scheduler/v1"
	storage "d7y.io/dragonfly/v2/client/daemon/storage"
	logger "d7y.io/dragonfly/v2/internal/dflog"
	gomock "go.uber.org/mock/gomock"
	status "google.golang.org/grpc/status"
)

// MockTaskManager is a mock of TaskManager interface.
type MockTaskManager struct {
	ctrl     *gomock.Controller
	recorder *MockTaskManagerMockRecorder
	isgomock struct{}
}

// MockTaskManagerMockRecorder is the mock recorder for MockTaskManager.
type MockTaskManagerMockRecorder struct {
	mock *MockTaskManager
}

// NewMockTaskManager creates a new mock instance.
func NewMockTaskManager(ctrl *gomock.Controller) *MockTaskManager {
	mock := &MockTaskManager{ctrl: ctrl}
	mock.recorder = &MockTaskManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTaskManager) EXPECT() *MockTaskManagerMockRecorder {
	return m.recorder
}

// AnnouncePeerTask mocks base method.
func (m *MockTaskManager) AnnouncePeerTask(ctx context.Context, meta storage.PeerTaskMetadata, url string, taskType common.TaskType, urlMeta *common.UrlMeta) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AnnouncePeerTask", ctx, meta, url, taskType, urlMeta)
	ret0, _ := ret[0].(error)
	return ret0
}

// AnnouncePeerTask indicates an expected call of AnnouncePeerTask.
func (mr *MockTaskManagerMockRecorder) AnnouncePeerTask(ctx, meta, url, taskType, urlMeta any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AnnouncePeerTask", reflect.TypeOf((*MockTaskManager)(nil).AnnouncePeerTask), ctx, meta, url, taskType, urlMeta)
}

// GetPieceManager mocks base method.
func (m *MockTaskManager) GetPieceManager() PieceManager {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPieceManager")
	ret0, _ := ret[0].(PieceManager)
	return ret0
}

// GetPieceManager indicates an expected call of GetPieceManager.
func (mr *MockTaskManagerMockRecorder) GetPieceManager() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPieceManager", reflect.TypeOf((*MockTaskManager)(nil).GetPieceManager))
}

// IsPeerTaskRunning mocks base method.
func (m *MockTaskManager) IsPeerTaskRunning(taskID, peerID string) (Task, bool) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsPeerTaskRunning", taskID, peerID)
	ret0, _ := ret[0].(Task)
	ret1, _ := ret[1].(bool)
	return ret0, ret1
}

// IsPeerTaskRunning indicates an expected call of IsPeerTaskRunning.
func (mr *MockTaskManagerMockRecorder) IsPeerTaskRunning(taskID, peerID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsPeerTaskRunning", reflect.TypeOf((*MockTaskManager)(nil).IsPeerTaskRunning), taskID, peerID)
}

// StartFileTask mocks base method.
func (m *MockTaskManager) StartFileTask(ctx context.Context, req *FileTaskRequest) (chan *FileTaskProgress, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StartFileTask", ctx, req)
	ret0, _ := ret[0].(chan *FileTaskProgress)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// StartFileTask indicates an expected call of StartFileTask.
func (mr *MockTaskManagerMockRecorder) StartFileTask(ctx, req any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StartFileTask", reflect.TypeOf((*MockTaskManager)(nil).StartFileTask), ctx, req)
}

// StartSeedTask mocks base method.
func (m *MockTaskManager) StartSeedTask(ctx context.Context, req *SeedTaskRequest) (*SeedTaskResponse, bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StartSeedTask", ctx, req)
	ret0, _ := ret[0].(*SeedTaskResponse)
	ret1, _ := ret[1].(bool)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// StartSeedTask indicates an expected call of StartSeedTask.
func (mr *MockTaskManagerMockRecorder) StartSeedTask(ctx, req any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StartSeedTask", reflect.TypeOf((*MockTaskManager)(nil).StartSeedTask), ctx, req)
}

// StartStreamTask mocks base method.
func (m *MockTaskManager) StartStreamTask(ctx context.Context, req *StreamTaskRequest) (io.ReadCloser, map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StartStreamTask", ctx, req)
	ret0, _ := ret[0].(io.ReadCloser)
	ret1, _ := ret[1].(map[string]string)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// StartStreamTask indicates an expected call of StartStreamTask.
func (mr *MockTaskManagerMockRecorder) StartStreamTask(ctx, req any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StartStreamTask", reflect.TypeOf((*MockTaskManager)(nil).StartStreamTask), ctx, req)
}

// StatTask mocks base method.
func (m *MockTaskManager) StatTask(ctx context.Context, taskID string) (*scheduler.Task, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StatTask", ctx, taskID)
	ret0, _ := ret[0].(*scheduler.Task)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// StatTask indicates an expected call of StatTask.
func (mr *MockTaskManagerMockRecorder) StatTask(ctx, taskID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StatTask", reflect.TypeOf((*MockTaskManager)(nil).StatTask), ctx, taskID)
}

// Stop mocks base method.
func (m *MockTaskManager) Stop(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Stop", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// Stop indicates an expected call of Stop.
func (mr *MockTaskManagerMockRecorder) Stop(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockTaskManager)(nil).Stop), ctx)
}

// Subscribe mocks base method.
func (m *MockTaskManager) Subscribe(request *common.PieceTaskRequest) (*SubscribeResponse, bool) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Subscribe", request)
	ret0, _ := ret[0].(*SubscribeResponse)
	ret1, _ := ret[1].(bool)
	return ret0, ret1
}

// Subscribe indicates an expected call of Subscribe.
func (mr *MockTaskManagerMockRecorder) Subscribe(request any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Subscribe", reflect.TypeOf((*MockTaskManager)(nil).Subscribe), request)
}

// MockTask is a mock of Task interface.
type MockTask struct {
	ctrl     *gomock.Controller
	recorder *MockTaskMockRecorder
	isgomock struct{}
}

// MockTaskMockRecorder is the mock recorder for MockTask.
type MockTaskMockRecorder struct {
	mock *MockTask
}

// NewMockTask creates a new mock instance.
func NewMockTask(ctrl *gomock.Controller) *MockTask {
	mock := &MockTask{ctrl: ctrl}
	mock.recorder = &MockTaskMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTask) EXPECT() *MockTaskMockRecorder {
	return m.recorder
}

// AddTraffic mocks base method.
func (m *MockTask) AddTraffic(arg0 uint64) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddTraffic", arg0)
}

// AddTraffic indicates an expected call of AddTraffic.
func (mr *MockTaskMockRecorder) AddTraffic(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddTraffic", reflect.TypeOf((*MockTask)(nil).AddTraffic), arg0)
}

// Context mocks base method.
func (m *MockTask) Context() context.Context {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Context")
	ret0, _ := ret[0].(context.Context)
	return ret0
}

// Context indicates an expected call of Context.
func (mr *MockTaskMockRecorder) Context() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Context", reflect.TypeOf((*MockTask)(nil).Context))
}

// GetContentLength mocks base method.
func (m *MockTask) GetContentLength() int64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetContentLength")
	ret0, _ := ret[0].(int64)
	return ret0
}

// GetContentLength indicates an expected call of GetContentLength.
func (mr *MockTaskMockRecorder) GetContentLength() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetContentLength", reflect.TypeOf((*MockTask)(nil).GetContentLength))
}

// GetPeerID mocks base method.
func (m *MockTask) GetPeerID() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPeerID")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetPeerID indicates an expected call of GetPeerID.
func (mr *MockTaskMockRecorder) GetPeerID() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPeerID", reflect.TypeOf((*MockTask)(nil).GetPeerID))
}

// GetPieceMd5Sign mocks base method.
func (m *MockTask) GetPieceMd5Sign() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPieceMd5Sign")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetPieceMd5Sign indicates an expected call of GetPieceMd5Sign.
func (mr *MockTaskMockRecorder) GetPieceMd5Sign() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPieceMd5Sign", reflect.TypeOf((*MockTask)(nil).GetPieceMd5Sign))
}

// GetStorage mocks base method.
func (m *MockTask) GetStorage() storage.TaskStorageDriver {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetStorage")
	ret0, _ := ret[0].(storage.TaskStorageDriver)
	return ret0
}

// GetStorage indicates an expected call of GetStorage.
func (mr *MockTaskMockRecorder) GetStorage() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStorage", reflect.TypeOf((*MockTask)(nil).GetStorage))
}

// GetTaskID mocks base method.
func (m *MockTask) GetTaskID() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTaskID")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetTaskID indicates an expected call of GetTaskID.
func (mr *MockTaskMockRecorder) GetTaskID() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTaskID", reflect.TypeOf((*MockTask)(nil).GetTaskID))
}

// GetTotalPieces mocks base method.
func (m *MockTask) GetTotalPieces() int32 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTotalPieces")
	ret0, _ := ret[0].(int32)
	return ret0
}

// GetTotalPieces indicates an expected call of GetTotalPieces.
func (mr *MockTaskMockRecorder) GetTotalPieces() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTotalPieces", reflect.TypeOf((*MockTask)(nil).GetTotalPieces))
}

// GetTraffic mocks base method.
func (m *MockTask) GetTraffic() uint64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTraffic")
	ret0, _ := ret[0].(uint64)
	return ret0
}

// GetTraffic indicates an expected call of GetTraffic.
func (mr *MockTaskMockRecorder) GetTraffic() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTraffic", reflect.TypeOf((*MockTask)(nil).GetTraffic))
}

// Log mocks base method.
func (m *MockTask) Log() *logger.SugaredLoggerOnWith {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Log")
	ret0, _ := ret[0].(*logger.SugaredLoggerOnWith)
	return ret0
}

// Log indicates an expected call of Log.
func (mr *MockTaskMockRecorder) Log() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Log", reflect.TypeOf((*MockTask)(nil).Log))
}

// PublishPieceInfo mocks base method.
func (m *MockTask) PublishPieceInfo(pieceNum int32, size uint32) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "PublishPieceInfo", pieceNum, size)
}

// PublishPieceInfo indicates an expected call of PublishPieceInfo.
func (mr *MockTaskMockRecorder) PublishPieceInfo(pieceNum, size any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PublishPieceInfo", reflect.TypeOf((*MockTask)(nil).PublishPieceInfo), pieceNum, size)
}

// ReportPieceResult mocks base method.
func (m *MockTask) ReportPieceResult(request *DownloadPieceRequest, result *DownloadPieceResult, err error) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "ReportPieceResult", request, result, err)
}

// ReportPieceResult indicates an expected call of ReportPieceResult.
func (mr *MockTaskMockRecorder) ReportPieceResult(request, result, err any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReportPieceResult", reflect.TypeOf((*MockTask)(nil).ReportPieceResult), request, result, err)
}

// SetContentLength mocks base method.
func (m *MockTask) SetContentLength(arg0 int64) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetContentLength", arg0)
}

// SetContentLength indicates an expected call of SetContentLength.
func (mr *MockTaskMockRecorder) SetContentLength(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetContentLength", reflect.TypeOf((*MockTask)(nil).SetContentLength), arg0)
}

// SetPieceMd5Sign mocks base method.
func (m *MockTask) SetPieceMd5Sign(arg0 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetPieceMd5Sign", arg0)
}

// SetPieceMd5Sign indicates an expected call of SetPieceMd5Sign.
func (mr *MockTaskMockRecorder) SetPieceMd5Sign(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetPieceMd5Sign", reflect.TypeOf((*MockTask)(nil).SetPieceMd5Sign), arg0)
}

// SetTotalPieces mocks base method.
func (m *MockTask) SetTotalPieces(arg0 int32) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetTotalPieces", arg0)
}

// SetTotalPieces indicates an expected call of SetTotalPieces.
func (mr *MockTaskMockRecorder) SetTotalPieces(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetTotalPieces", reflect.TypeOf((*MockTask)(nil).SetTotalPieces), arg0)
}

// UpdateSourceErrorStatus mocks base method.
func (m *MockTask) UpdateSourceErrorStatus(st *status.Status) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "UpdateSourceErrorStatus", st)
}

// UpdateSourceErrorStatus indicates an expected call of UpdateSourceErrorStatus.
func (mr *MockTaskMockRecorder) UpdateSourceErrorStatus(st any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateSourceErrorStatus", reflect.TypeOf((*MockTask)(nil).UpdateSourceErrorStatus), st)
}

// MockLogger is a mock of Logger interface.
type MockLogger struct {
	ctrl     *gomock.Controller
	recorder *MockLoggerMockRecorder
	isgomock struct{}
}

// MockLoggerMockRecorder is the mock recorder for MockLogger.
type MockLoggerMockRecorder struct {
	mock *MockLogger
}

// NewMockLogger creates a new mock instance.
func NewMockLogger(ctrl *gomock.Controller) *MockLogger {
	mock := &MockLogger{ctrl: ctrl}
	mock.recorder = &MockLoggerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockLogger) EXPECT() *MockLoggerMockRecorder {
	return m.recorder
}

// Log mocks base method.
func (m *MockLogger) Log() *logger.SugaredLoggerOnWith {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Log")
	ret0, _ := ret[0].(*logger.SugaredLoggerOnWith)
	return ret0
}

// Log indicates an expected call of Log.
func (mr *MockLoggerMockRecorder) Log() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Log", reflect.TypeOf((*MockLogger)(nil).Log))
}
