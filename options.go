package grpc_sentry

import (
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	defaultServerOperationName = "grpc.server"
	defaultClientOperationName = "grpc.client"
)

var defaultOptions = &options{
	Repanic:               false,
	WaitForDelivery:       false,
	ReportOn:              ReportAlways,
	Timeout:               1 * time.Second,
	OperationNameOverride: "",
	CaptureRequestBody:    true,
}

type options struct {
	// Repanic configures whether Sentry should repanic after recovery.
	Repanic bool

	// WaitForDelivery configures whether you want to block the request before moving forward with the response.
	WaitForDelivery bool

	// Timeout for the event delivery requests.
	Timeout time.Duration

	ReportOn func(error) bool

	OperationNameOverride string

	// CaptureRequestBody configures whether the request body should be sent to Sentry.
	CaptureRequestBody bool
}

func ReportAlways(error) bool {
	return true
}

func ReportOnCodes(cc ...codes.Code) reporter {
	return func(err error) bool {
		for i := range cc {
			if status.Code(err) == cc[i] {
				return true
			}
		}
		return false
	}
}
