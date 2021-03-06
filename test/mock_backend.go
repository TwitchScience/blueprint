package test

import (
	"errors"
	"sync"

	"github.com/twitchscience/blueprint/bpdb"
	"github.com/twitchscience/blueprint/core"
	"github.com/twitchscience/scoop_protocol/scoop_protocol"
)

// MockBpdb is a mock for the bpdb/Bpdb interface which tracks whether the DB is in maintenance mode.
type MockBpdb struct {
	maintenanceMutex *sync.RWMutex
	maintenanceMode  bpdb.MaintenanceMode
	maintenanceModes map[string]bpdb.MaintenanceMode
	mockActiveUsers  []*bpdb.ActiveUser
	mockDailyChanges []*bpdb.DailyChange
}

// MockBpSchemaBackend is a mock for the bpdb/BpSchemaBackend interface which tracks how many times AllSchemas has been called
type MockBpSchemaBackend struct {
	allSchemasMutex       *sync.RWMutex
	allEventMetadataMutex *sync.RWMutex
	allSchemasCalls       int32
	allEventMetadataCalls int32
	metadataState         map[string](map[string]bpdb.EventMetadataRow)
}

// MockBpKinesisConfigBackend is a mock for the bpdb/BpKinesisConfigBackend interface
type MockBpKinesisConfigBackend struct {
}

// NewMockBpdb creates a new mock backend.
func NewMockBpdb(mm map[string]bpdb.MaintenanceMode, activeUsers []*bpdb.ActiveUser, dailyChanges []*bpdb.DailyChange) *MockBpdb {
	return &MockBpdb{&sync.RWMutex{}, bpdb.MaintenanceMode{IsInMaintenanceMode: false, User: ""}, mm, activeUsers, dailyChanges}
}

// NewMockBpSchemaBackend creates a new mock schema backend.
func NewMockBpSchemaBackend(initMetadata map[string]map[string]bpdb.EventMetadataRow) *MockBpSchemaBackend {
	return &MockBpSchemaBackend{&sync.RWMutex{}, &sync.RWMutex{}, 0, 0, initMetadata}
}

// NewMockBpKinesisConfigBackend creates a new mock kinesis config backend.
func NewMockBpKinesisConfigBackend() *MockBpKinesisConfigBackend {
	return &MockBpKinesisConfigBackend{}
}

// GetAllSchemasCalls returns the number of times AllSchemas() has been called.
func (m *MockBpSchemaBackend) GetAllSchemasCalls() int32 {
	m.allSchemasMutex.RLock()
	defer m.allSchemasMutex.RUnlock()
	return m.allSchemasCalls
}

// AllSchemas increments the number of AllSchemas calls and return nils.
func (m *MockBpSchemaBackend) AllSchemas() ([]bpdb.AnnotatedSchema, error) {
	m.allSchemasMutex.Lock()
	m.allSchemasCalls++
	m.allSchemasMutex.Unlock()
	return make([]bpdb.AnnotatedSchema, 0), nil
}

// Schema returns nils except when the event name is "this-table-exists" or "this-event-exists"
func (m *MockBpSchemaBackend) Schema(name string, version *int) (*bpdb.AnnotatedSchema, error) {
	if name == "this-table-exists" || name == "this-event-exists" {
		return &bpdb.AnnotatedSchema{}, nil
	}
	return nil, nil
}

// UpdateSchema returns nil.
func (m *MockBpSchemaBackend) UpdateSchema(update *core.ClientUpdateSchemaRequest, user string) *core.WebError {
	return nil
}

// CreateSchema returns nil.
func (m *MockBpSchemaBackend) CreateSchema(schema *scoop_protocol.Config, user string) *core.WebError {
	return nil
}

// Migration returns nils.
func (m *MockBpSchemaBackend) Migration(table string, from int, to int) ([]*scoop_protocol.Operation, error) {
	return nil, nil
}

// DropSchema return nil.
func (m *MockBpSchemaBackend) DropSchema(schema *bpdb.AnnotatedSchema, reason string, exists bool, user string) error {
	return nil
}

// AllEventMetadata increments the number of AllEventMetadata calls
func (m *MockBpSchemaBackend) AllEventMetadata() (*bpdb.AllEventMetadata, error) {
	m.allEventMetadataMutex.Lock()
	m.allEventMetadataCalls++
	m.allEventMetadataMutex.Unlock()
	return &bpdb.AllEventMetadata{Metadata: m.metadataState}, nil
}

// GetAllEventMetadataCalls returns the number of times EventMetadata() has been called.
func (m *MockBpSchemaBackend) GetAllEventMetadataCalls() int32 {
	m.allEventMetadataMutex.RLock()
	defer m.allEventMetadataMutex.RUnlock()
	return m.allEventMetadataCalls
}

// UpdateEventMetadata returns nil if update.EventName is in the returnMap
func (m *MockBpSchemaBackend) UpdateEventMetadata(update *core.ClientUpdateEventMetadataRequest, user string) *core.WebError {
	if _, exists := m.metadataState[update.EventName]; exists {
		m.metadataState[update.EventName][string(update.MetadataType)] = bpdb.EventMetadataRow{
			MetadataValue: update.MetadataValue,
		}
		return nil
	}
	return core.NewUserWebError(errors.New("schema does not exist"))
}

// AllKinesisConfigs returns nil
func (m *MockBpKinesisConfigBackend) AllKinesisConfigs() ([]scoop_protocol.AnnotatedKinesisConfig, error) {
	return make([]scoop_protocol.AnnotatedKinesisConfig, 0), nil
}

// KinesisConfig returns nil
func (m *MockBpKinesisConfigBackend) KinesisConfig(account int64, streamType string, name string) (*scoop_protocol.AnnotatedKinesisConfig, error) {
	return nil, nil
}

// UpdateKinesisConfig returns nil
func (m *MockBpKinesisConfigBackend) UpdateKinesisConfig(update *scoop_protocol.AnnotatedKinesisConfig, user string) *core.WebError {
	return nil
}

// CreateKinesisConfig returns nil
func (m *MockBpKinesisConfigBackend) CreateKinesisConfig(config *scoop_protocol.AnnotatedKinesisConfig, user string) *core.WebError {
	return nil
}

// DropKinesisConfig returns nil
func (m *MockBpKinesisConfigBackend) DropKinesisConfig(config *scoop_protocol.AnnotatedKinesisConfig, reason string, user string) error {
	return nil
}

// GetMaintenanceMode returns current value (starts as false, can be set by SetMaintenanceMode).
func (m *MockBpdb) GetMaintenanceMode() bpdb.MaintenanceMode {
	m.maintenanceMutex.RLock()
	defer m.maintenanceMutex.RUnlock()
	return m.maintenanceMode
}

// SetMaintenanceMode sets the maintenance mode in memory and returns nil.
func (m *MockBpdb) SetMaintenanceMode(switchingOn bool, user, reason string) error {
	m.maintenanceMutex.Lock()
	m.maintenanceMode = bpdb.MaintenanceMode{IsInMaintenanceMode: switchingOn, User: user}
	m.maintenanceMutex.Unlock()
	return nil
}

// ActiveUsersLast30Days returns nils.
func (m *MockBpdb) ActiveUsersLast30Days() ([]*bpdb.ActiveUser, error) {
	return m.mockActiveUsers, nil
}

// DailyChangesLast30Days returns nils.
func (m *MockBpdb) DailyChangesLast30Days() ([]*bpdb.DailyChange, error) {
	return m.mockDailyChanges, nil
}

// GetSchemaMaintenanceMode returns false, ""
func (m *MockBpdb) GetSchemaMaintenanceMode(schema string) (bpdb.MaintenanceMode, error) {
	mm, e := m.maintenanceModes[schema]
	if !e {
		return bpdb.MaintenanceMode{IsInMaintenanceMode: false, User: ""}, nil
	}
	return mm, nil
}

// SetSchemaMaintenanceMode returns nil
func (m *MockBpdb) SetSchemaMaintenanceMode(schema string, switchingOn bool, user, reason string) error {
	return nil
}
