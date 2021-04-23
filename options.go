package grpc_sentry

import (
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var defaultOptions = &options{
		Repanic: false,
		WaitForDelivery: false,
		ReportOn: ReportAlways,
		Timeout: 1 * time.Second,
}


type options struct {
	// Repanic configures whether Sentry should repanic after recovery.
	Repanic bool

	// WaitForDelivery configures whether you want to block the request before moving forward with the response.
	WaitForDelivery bool

	// Timeout for the event delivery requests.
	Timeout time.Duration

	ReportOn func(error) bool
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
