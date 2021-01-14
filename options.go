package grpc_sentry

import (
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	defaultServerOptions = &serveroptions{
		Repanic: false,
		WaitForDelivery: false,
		ReportOn: ReportAlways,
		Timeout: 1 * time.Second,
	}

	defaultClientOptions = &clientoptions{
		Repanic: false,
		WaitForDelivery: false,
		ReportOn: ReportAlways,
		Timeout: 1 * time.Second,
	}
)

type options struct {
	// Repanic configures whether Sentry should repanic after recovery.
	Repanic bool

	// WaitForDelivery configures whether you want to block the request before moving forward with the response.
	WaitForDelivery bool

	// Timeout for the event delivery requests.
	Timeout time.Duration

	ReportOn func(error) bool
}

type clientoptions options

type serveroptions options

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
