package e

import (
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Status struct {
	*status.Status
}

// Error new status with code and message
func Error(code codes.Code, message string) *Status {
	return &Status{status.New(code, message)}
}

// Errorf new status with code and message
func Errorf(code codes.Code, format string, args ...interface{}) *Status {
	return &Status{status.Newf(code, format, args...)}
}

// GRPCStatus ...
func (s Status) GRPCStatus() *status.Status {
	return s.Status
}

// Error ...
func (s Status) Error() string {
	if m := s.Message(); m != "" {
		return m
	}

	return strconv.Itoa(int(s.Code()))
}
