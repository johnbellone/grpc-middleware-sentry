// SPDX-License-Identifier: Apache-2.0
package grpc_sentry

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// mockUnaryInvoker is a mock invoker for testing unary client interceptors
type mockUnaryInvoker struct {
	response interface{}
	err      error
}

func (m *mockUnaryInvoker) invoke(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
	// Simulate setting the reply
	if reply != nil && m.response != nil {
		// This is a simplified mock - in real usage, you'd need to handle type conversion
		// For testing purposes, we'll just check if the call was made
	}
	return m.err
}

// mockStreamer is a mock streamer for testing stream client interceptors
type mockStreamer struct {
	clientStream grpc.ClientStream
	err          error
}

func (m *mockStreamer) stream(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return m.clientStream, m.err
}

// mockClientStream implements grpc.ClientStream for testing
type mockClientStream struct {
	ctx context.Context
}

func (m *mockClientStream) Header() (metadata.MD, error) { return nil, nil }
func (m *mockClientStream) Trailer() metadata.MD         { return nil }
func (m *mockClientStream) CloseSend() error             { return nil }
func (m *mockClientStream) Context() context.Context     { return m.ctx }
func (m *mockClientStream) SendMsg(interface{}) error    { return nil }
func (m *mockClientStream) RecvMsg(interface{}) error    { return nil }

func TestUnaryClientInterceptor_Configuration(t *testing.T) {
	// Test that the interceptor can be created with different options
	tests := []struct {
		name    string
		options []Option
	}{
		{
			name:    "default options",
			options: []Option{},
		},
		{
			name:    "with operation name override",
			options: []Option{WithOperationNameOverride("custom.client")},
		},
		{
			name:    "with timeout",
			options: []Option{WithTimeout(5 * time.Second)},
		},
		{
			name:    "with custom report function",
			options: []Option{WithReportOn(ReportOnCodes(codes.InvalidArgument))},
		},
		{
			name: "multiple options",
			options: []Option{
				WithOperationNameOverride("test.client"),
				WithTimeout(10 * time.Second),
				WithReportOn(ReportOnCodes(codes.Internal, codes.NotFound)),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This should not panic
			interceptor := UnaryClientInterceptor(tt.options...)
			if interceptor == nil {
				t.Error("Expected interceptor to be created, got nil")
			}
		})
	}
}

func TestStreamClientInterceptor_Configuration(t *testing.T) {
	// Test that the interceptor can be created with different options
	tests := []struct {
		name    string
		options []Option
	}{
		{
			name:    "default options",
			options: []Option{},
		},
		{
			name:    "with operation name override",
			options: []Option{WithOperationNameOverride("custom.client.stream")},
		},
		{
			name:    "with timeout",
			options: []Option{WithTimeout(5 * time.Second)},
		},
		{
			name:    "with custom report function",
			options: []Option{WithReportOn(ReportOnCodes(codes.Unavailable))},
		},
		{
			name: "multiple options",
			options: []Option{
				WithOperationNameOverride("test.client.stream"),
				WithTimeout(10 * time.Second),
				WithReportOn(ReportOnCodes(codes.Internal, codes.NotFound)),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This should not panic
			interceptor := StreamClientInterceptor(tt.options...)
			if interceptor == nil {
				t.Error("Expected interceptor to be created, got nil")
			}
		})
	}
}

func TestReportOnCodes_Client(t *testing.T) {
	tests := []struct {
		name        string
		error       error
		codes       []codes.Code
		shouldReport bool
	}{
		{
			name:        "report on internal error",
			error:       status.Error(codes.Internal, "internal error"),
			codes:       []codes.Code{codes.Internal},
			shouldReport: true,
		},
		{
			name:        "don't report on different error",
			error:       status.Error(codes.InvalidArgument, "invalid argument"),
			codes:       []codes.Code{codes.Internal},
			shouldReport: false,
		},
		{
			name:        "report on multiple codes",
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

func TestReportAlways_Client(t *testing.T) {
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
