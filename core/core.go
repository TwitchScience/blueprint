package core

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/twitchscience/aws_utils/logger"
	"github.com/twitchscience/scoop_protocol/scoop_protocol"
)

// Subprocess represents something that can be set up, started, and stopped. E.g. a server.
type Subprocess interface {
	Setup() error
	Start()
	Stop()
}

// SubprocessManager oversees multiple subprocesses.
type SubprocessManager struct {
	Processes []Subprocess
	wg        *sync.WaitGroup
}

// Start all subprocesses. If any return an error, exits the process.
func (s *SubprocessManager) Start() {
	var wg sync.WaitGroup
	for _, sp := range s.Processes {
		err := sp.Setup()
		if err != nil {
			logger.WithError(err).WithField("subprocess", sp).Fatal("Failed to set up subprocess")
		}

		fn := sp
		wg.Add(1)
		logger.Go(func() {
			fn.Start()
			wg.Done()
		})
	}
	s.wg = &wg
}

// Wait for all subprocesses to finish, i.e. return from their Start method.
func (s *SubprocessManager) Wait() {
	s.wg.Wait()
}

// Stop all subprocesses.
func (s *SubprocessManager) Stop() {
	for _, sp := range s.Processes {
		sp.Stop()
	}
}

// Column represents a SQL column in a table for event data.
type Column struct {
	// InboundName is the name of the event property.
	InboundName string `json:"InboundName"`

	// OutboundName is the name of the column for the event property.
	OutboundName string `json:"OutboundName"`

	// Transformer is the column's SQL type.
	Transformer string `json:"Transformer"`

	// Length is the length of the SQL type, e.g. for a variable type like varchar.
	// TODO: length should be an int, currently the client supplies this
	// to us, so pass through now, with a view to fixing this later
	Length string `json:"ColumnCreationOptions"`

	// SupportingColumns are the names of extra columns required to map a value to this column
	SupportingColumns string `json:"SupportingColumns"`
}

// Renames is a map of old name to new name, representing a rename operation on
// a set of columns.
type Renames map[string]string

// ClientUpdateSchemaRequest is a request to update the schema for an event.
type ClientUpdateSchemaRequest struct {
	EventName string `json:"-"`
	Additions []Column
	Deletes   []string
	Renames   Renames
}

// ClientDropSchemaRequest is a request to drop the schema for an event.
type ClientDropSchemaRequest struct {
	EventName string
	Reason    string
}

// ClientUpdateEventCommentRequest is a request to update the comment for an event.
type ClientUpdateEventCommentRequest struct {
	EventName    string
	EventComment string
}

// ClientUpdateEventMetadataRequest is a request to update the metadata for an event.
type ClientUpdateEventMetadataRequest struct {
	EventName     string
	MetadataType  scoop_protocol.EventMetadataType
	MetadataValue string
}

// WebError is either a server or user error.
type WebError struct {
	ServerError error
	UserError   error
}

// ReportError reports the WebError's error and the given message to the ResponseWriter/logger.
func (we WebError) ReportError(w http.ResponseWriter, message string) {
	if we.ServerError != nil {
		logger.WithError(we.ServerError).Error(message)
		http.Error(w, "Internal error: "+message, http.StatusInternalServerError)
	} else if we.UserError != nil {
		logger.WithError(we.UserError).Info(message)
		http.Error(w, message+": "+we.UserError.Error(), http.StatusBadRequest)
	}
}

// NewServerWebError returns a WebError representing a server error.
func NewServerWebError(err error) *WebError {
	if err == nil {
		return nil
	}
	return &WebError{ServerError: err}
}

// NewServerWebErrorf formats a WebError representing a server error.
func NewServerWebErrorf(format string, a ...interface{}) *WebError {
	return &WebError{ServerError: fmt.Errorf(format, a...)}
}

// NewUserWebError returns a WebError representing a user error.
func NewUserWebError(err error) *WebError {
	if err == nil {
		return nil
	}
	return &WebError{UserError: err}
}

// NewUserWebErrorf formats a WebError representing a user error.
func NewUserWebErrorf(format string, a ...interface{}) *WebError {
	return &WebError{UserError: fmt.Errorf(format, a...)}
}

// AnnotateWebError adds a prefix to the error string for the web error, colon
// delimited
func AnnotateWebError(msg string, err *WebError) *WebError {
	if err.UserError != nil {
		return &WebError{UserError: fmt.Errorf(msg+": %v", err.UserError)}
	}
	return &WebError{ServerError: fmt.Errorf(msg+": %v", err.ServerError)}

}
