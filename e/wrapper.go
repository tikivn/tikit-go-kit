package e

import (
	"google.golang.org/grpc/codes"
)

const (
	clientErrMsg = "bad request"
)

// WrapError ...
func WrapError(err error) *Status {
	if stt, ok := err.(*Status); ok {
		return stt
	}

	return Error(codes.InvalidArgument, clientErrMsg)
}

// WrapErrorf ...
func WrapErrorf(err error, format string, args ...interface{}) *Status {
	if stt, ok := err.(*Status); ok {
		return stt
	}

	return Errorf(codes.InvalidArgument, format, args...)
}
