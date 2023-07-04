package osutils_api

import (
	"github.com/portworx/torpedo/pkg/osutils"
)

const (
	// DefaultExecShellCommand is the default command for ExecShell
	DefaultExecShellCommand = "hostname"
)

const (
	// DefaultExecShellStdOut is the default std-out for ExecShell
	DefaultExecShellStdOut = ""
	// DefaultExecShellStdErr is the default std-err for ExecShell
	DefaultExecShellStdErr = ""
)

// ExecShellRequest represents a backup_request_manager.Request for the osutils.ExecShell
type ExecShellRequest struct {
	Command string
}

// GetCommand returns the Command associated with the ExecShellRequest
func (r *ExecShellRequest) GetCommand() string {
	return r.Command
}

// SetCommand sets the Command for the ExecShellRequest
func (r *ExecShellRequest) SetCommand(command string) *ExecShellRequest {
	r.Command = command
	return r
}

// NewExecShellRequest creates a new instance of the ExecShellRequest
func NewExecShellRequest(command string) *ExecShellRequest {
	newExecShellRequest := &ExecShellRequest{}
	newExecShellRequest.SetCommand(command)
	return newExecShellRequest
}

// NewDefaultExecShellRequest creates a new instance of the ExecShellRequest with default values
func NewDefaultExecShellRequest() *ExecShellRequest {
	return NewExecShellRequest(DefaultExecShellCommand)
}

// ExecShellResponse represents a backup_request_manager.Response for the osutils.ExecShell
type ExecShellResponse struct {
	StdOut string
	StdErr string
}

// GetStdOut returns the StdOut associated with the ExecShellResponse
func (r *ExecShellResponse) GetStdOut() string {
	return r.StdOut
}

// SetStdOut sets the StdOut for the ExecShellResponse
func (r *ExecShellResponse) SetStdOut(stdOut string) *ExecShellResponse {
	r.StdOut = stdOut
	return r
}

// GetStdErr returns the StdErr associated with the ExecShellResponse
func (r *ExecShellResponse) GetStdErr() string {
	return r.StdErr
}

// SetStdErr sets the StdErr for the ExecShellResponse
func (r *ExecShellResponse) SetStdErr(stdErr string) *ExecShellResponse {
	r.StdErr = stdErr
	return r
}

// NewExecShellResponse creates a new instance of the ExecShellResponse
func NewExecShellResponse(stdOut string, stdErr string) *ExecShellResponse {
	newExecShellResponse := &ExecShellResponse{}
	newExecShellResponse.SetStdOut(stdOut)
	newExecShellResponse.SetStdErr(stdErr)
	return newExecShellResponse
}

// NewDefaultExecShellResponse creates a new instance of the ExecShellResponse with default values
func NewDefaultExecShellResponse() *ExecShellResponse {
	return NewExecShellResponse(DefaultExecShellStdOut, DefaultExecShellStdErr)
}

// ExecShell executes shell command
func ExecShell(request *ExecShellRequest) (*ExecShellResponse, error) {
	stdout, stderr, err := osutils.ExecShell(request.GetCommand())
	if err != nil {
		return nil, err
	}
	return NewExecShellResponse(stdout, stderr), nil
}
