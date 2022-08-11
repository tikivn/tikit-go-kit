package e

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	clientErrStatus = http.StatusBadRequest
	serverErrStatus = http.StatusInternalServerError
)

type Status struct {
	HTTPStatus int
	Err        *status.Status
}

// Error new status with code and message
func Error(code codes.Code, message string) *Status {
	return &Status{
		HTTPStatus: clientErrStatus,
		Err:        status.New(code, message)}
}

// Errorf new status with code and message
func Errorf(code codes.Code, format string, args ...interface{}) *Status {
	return &Status{
		HTTPStatus: clientErrStatus,
		Err:        status.Newf(code, format, args...)}
}

// SetHttpStatus set http status
func (s *Status) SetHttpStatus(code int) error {
	s.HTTPStatus = code
	return s
}

// ClientErr http status 400 - bad request
func (s *Status) ClientErr() error {
	return s.SetHttpStatus(clientErrStatus)
}

// ServerErr http status 500 - internal server error
func (s *Status) ServerErr() error {
	return s.SetHttpStatus(serverErrStatus)
}

// GRPCStatus ...
func (s Status) GRPCStatus() *status.Status {
	return s.Err
}

// Code ...
func (s Status) Code() codes.Code {
	return s.Err.Code()
}

// Message ...
func (s Status) Message() string {
	return s.Err.Message()
}

// Error ...
func (s *Status) Error() string {
	if m := s.Message(); m != "" {
		return m
	}

	return strconv.Itoa(int(s.Err.Code()))
}

// Unwrap ...
func (s *Status) Unwrap() error {
	return fmt.Errorf("%w", &runtime.HTTPStatusError{
		HTTPStatus: s.HTTPStatus,
		Err:        s.Err.Err(),
	})
}
