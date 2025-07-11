// SPDX-License-Identifier: Apache-2.0
package grpc_sentry

import (
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestReportAlways(t *testing.T) {
	// Test that ReportAlways always returns true regardless of the error
	tests := []struct {
		name  string
		error error
	}{
		{
			name:  "nil error",
			error: nil,
		},
		{
			name:  "status error",
			error: status.Error(codes.Internal, "internal error"),
		},
		{
			name:  "different status error",
			error: status.Error(codes.InvalidArgument, "invalid argument"),
		},
		{
			name:  "non-status error",
			error: &testError{message: "test error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ReportAlways(tt.error)
			if result != true {
				t.Errorf("Expected ReportAlways to return true for %v, got %v", tt.error, result)
			}
		})
	}
}

func TestReportOnCodes(t *testing.T) {
	tests := []struct {
		name        string
		error       error
		codes       []codes.Code
		shouldReport bool
	}{
		{
			name:        "report on matching code",
			error:       status.Error(codes.Internal, "internal error"),
			codes:       []codes.Code{codes.Internal},
			shouldReport: true,
		},
		{
			name:        "don't report on non-matching code",
			error:       status.Error(codes.InvalidArgument, "invalid argument"),
			codes:       []codes.Code{codes.Internal},
			shouldReport: false,
		},
		{
			name:        "report on one of multiple codes",
			error:       status.Error(codes.NotFound, "not found"),
			codes:       []codes.Code{codes.Internal, codes.NotFound, codes.Unavailable},
			shouldReport: true,
		},
		{
			name:        "don't report on none of multiple codes",
			error:       status.Error(codes.InvalidArgument, "invalid argument"),
			codes:       []codes.Code{codes.Internal, codes.NotFound, codes.Unavailable},
			shouldReport: false,
		},
		{
			name:        "nil error with any codes",
			error:       nil,
			codes:       []codes.Code{codes.Internal, codes.NotFound},
			shouldReport: false,
		},
		{
			name:        "non-status error with any codes",
			error:       &testError{message: "test error"},
			codes:       []codes.Code{codes.Internal, codes.NotFound},
			shouldReport: false,
		},
		{
			name:        "empty codes list",
			error:       status.Error(codes.Internal, "internal error"),
			codes:       []codes.Code{},
			shouldReport: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := ReportOnCodes(tt.codes...)
			result := reporter(tt.error)

			if result != tt.shouldReport {
				t.Errorf("Expected ReportOnCodes to return %v for error %v with codes %v, got %v",
					tt.shouldReport, tt.error, tt.codes, result)
			}
		})
	}
}

func TestReportOnCodes_AllCodes(t *testing.T) {
	// Test all gRPC status codes to ensure they work correctly
	allCodes := []codes.Code{
		codes.OK,
		codes.Canceled,
		codes.Unknown,
		codes.InvalidArgument,
		codes.DeadlineExceeded,
		codes.NotFound,
		codes.AlreadyExists,
		codes.PermissionDenied,
		codes.ResourceExhausted,
		codes.FailedPrecondition,
		codes.Aborted,
		codes.OutOfRange,
		codes.Unimplemented,
		codes.Internal,
		codes.Unavailable,
		codes.DataLoss,
		codes.Unauthenticated,
	}

	for _, code := range allCodes {
		t.Run(code.String(), func(t *testing.T) {
			reporter := ReportOnCodes(code)
			error := status.Error(code, "test error")
			result := reporter(error)

			if result != true {
				t.Errorf("Expected ReportOnCodes to return true for code %s, got %v", code, result)
			}
		})
	}
}

func TestReportOnCodes_Combination(t *testing.T) {
	// Test combinations of codes
	tests := []struct {
		name        string
		error       error
		codes       []codes.Code
		shouldReport bool
	}{
		{
			name:        "multiple matching codes",
			error:       status.Error(codes.Internal, "internal error"),
			codes:       []codes.Code{codes.Internal, codes.Internal, codes.Internal},
			shouldReport: true,
		},
		{
			name:        "mixed matching and non-matching codes",
			error:       status.Error(codes.NotFound, "not found"),
			codes:       []codes.Code{codes.Internal, codes.NotFound, codes.Unavailable},
			shouldReport: true,
		},
		{
			name:        "all non-matching codes",
			error:       status.Error(codes.InvalidArgument, "invalid argument"),
			codes:       []codes.Code{codes.Internal, codes.NotFound, codes.Unavailable},
			shouldReport: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := ReportOnCodes(tt.codes...)
			result := reporter(tt.error)

			if result != tt.shouldReport {
				t.Errorf("Expected ReportOnCodes to return %v for error %v with codes %v, got %v",
					tt.shouldReport, tt.error, tt.codes, result)
			}
		})
	}
}

// testError is a simple error type for testing
type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}

func TestDefaultOptions(t *testing.T) {
	// Test that default options are set correctly
	if defaultOptions.Repanic != false {
		t.Errorf("Expected default Repanic to be false, got %v", defaultOptions.Repanic)
	}
	if defaultOptions.WaitForDelivery != false {
		t.Errorf("Expected default WaitForDelivery to be false, got %v", defaultOptions.WaitForDelivery)
	}
	if defaultOptions.Timeout != 1*time.Second {
		t.Errorf("Expected default Timeout to be 1s, got %v", defaultOptions.Timeout)
	}
	if defaultOptions.OperationNameOverride != "" {
		t.Errorf("Expected default OperationNameOverride to be empty, got %v", defaultOptions.OperationNameOverride)
	}
	if defaultOptions.CaptureRequestBody != true {
		t.Errorf("Expected default CaptureRequestBody to be true, got %v", defaultOptions.CaptureRequestBody)
	}
	if defaultOptions.ReportOn == nil {
		t.Error("Expected default ReportOn to be set, got nil")
	}
}

func TestDefaultOperationNames(t *testing.T) {
	// Test that default operation names are set correctly
	if defaultServerOperationName != "grpc.server" {
		t.Errorf("Expected defaultServerOperationName to be 'grpc.server', got %s", defaultServerOperationName)
	}
	if defaultClientOperationName != "grpc.client" {
		t.Errorf("Expected defaultClientOperationName to be 'grpc.client', got %s", defaultClientOperationName)
	}
}
