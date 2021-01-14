package grpc_sentry

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	defaultServerOptions = &serveroptions{
		Repanic: false,
		ReportOn: ReportAlways,
	}

	defaultClientOptions = &clientoptions{}
)

type clientoptions struct {
	ReportOn func(error) bool
}

type serveroptions struct {
	Repanic bool

	ReportOn func(error) bool
}

func evaluateClientOptions(opts []ClientOption) *clientoptions {
	optCopy := &clientoptions{}
	*optCopy = *defaultClientOptions
	for _, o := range opts {
		o(optCopy)
	}
	return optCopy
}

func evaluateServerOptions(opts []ServerOption) *serveroptions {
	optCopy := &serveroptions{}
	*optCopy = *defaultServerOptions
	for _, o := range opts {
		o(optCopy)
	}
	return optCopy
}

type ClientOption func(*clientoptions)
type ServerOption func(*serveroptions)

func ReportAlways(error) bool {
	return true
}

func ReportOnCodes(cc ...codes.Code) func(error) bool {
	return func(err error) bool {
		for i := range cc {
			if status.Code(err) == cc[i] {
				return true
			}
		}
		return false
	}
}
