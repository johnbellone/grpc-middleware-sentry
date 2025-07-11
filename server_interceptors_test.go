// SPDX-License-Identifier: Apache-2.0
package grpc_sentry

import (
	"context"
	"testing"
	"time"

	"github.com/getsentry/sentry-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

// mockUnaryHandler is a mock handler for testing unary interceptors
type mockUnaryHandler struct {
	response interface{}
	err      error
	panic    bool
}

func (h *mockUnaryHandler) handle(ctx context.Context, req interface{}) (interface{}, error) {
	if h.panic {
		panic("test panic")
	}
	return h.response, h.err
}

// mockStreamHandler is a mock handler for testing stream interceptors
type mockStreamHandler struct {
	err   error
	panic bool
}

func (h *mockStreamHandler) handle(srv interface{}, stream grpc.ServerStream) error {
	if h.panic {
		panic("test panic")
	}
	return h.err
}

// mockServerStream implements grpc.ServerStream for testing
type mockServerStream struct {
	ctx context.Context
}

func (m *mockServerStream) SetHeader(metadata.MD) error   { return nil }
func (m *mockServerStream) SendHeader(metadata.MD) error  { return nil }
func (m *mockServerStream) SetTrailer(metadata.MD)        {}
func (m *mockServerStream) Context() context.Context      { return m.ctx }
func (m *mockServerStream) SendMsg(interface{}) error     { return nil }
func (m *mockServerStream) RecvMsg(interface{}) error     { return nil }

func TestUnaryServerInterceptor_Configuration(t *testing.T) {
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
			name:    "with repanic",
			options: []Option{WithRepanicOption(true)},
		},
		{
			name:    "with wait for delivery",
			options: []Option{WithWaitForDelivery(true)},
		},
		{
			name:    "with timeout",
			options: []Option{WithTimeout(5 * time.Second)},
		},
		{
			name:    "with operation name override",
			options: []Option{WithOperationNameOverride("custom.operation")},
		},
		{
			name:    "with capture request body disabled",
			options: []Option{WithCaptureRequestBody(false)},
		},
		{
			name:    "with custom report function",
			options: []Option{WithReportOn(ReportOnCodes(codes.Internal))},
		},
		{
			name: "multiple options",
			options: []Option{
				WithRepanicOption(true),
				WithWaitForDelivery(true),
				WithTimeout(10 * time.Second),
				WithOperationNameOverride("test.operation"),
				WithCaptureRequestBody(false),
				WithReportOn(ReportOnCodes(codes.Internal, codes.NotFound)),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This should not panic
			interceptor := UnaryServerInterceptor(tt.options...)
			if interceptor == nil {
				t.Error("Expected interceptor to be created, got nil")
			}
		})
	}
}

func TestStreamServerInterceptor_Configuration(t *testing.T) {
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
			name:    "with repanic",
			options: []Option{WithRepanicOption(true)},
		},
		{
			name:    "with wait for delivery",
			options: []Option{WithWaitForDelivery(true)},
		},
		{
			name:    "with timeout",
			options: []Option{WithTimeout(5 * time.Second)},
		},
		{
			name:    "with operation name override",
			options: []Option{WithOperationNameOverride("custom.stream")},
		},
		{
			name:    "with custom report function",
			options: []Option{WithReportOn(ReportOnCodes(codes.Internal))},
		},
		{
			name: "multiple options",
			options: []Option{
				WithRepanicOption(true),
				WithWaitForDelivery(true),
				WithTimeout(10 * time.Second),
				WithOperationNameOverride("test.stream"),
				WithReportOn(ReportOnCodes(codes.Internal, codes.NotFound)),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This should not panic
			interceptor := StreamServerInterceptor(tt.options...)
			if interceptor == nil {
				t.Error("Expected interceptor to be created, got nil")
			}
		})
	}
}

func TestContinueFromGrpcMetadata(t *testing.T) {
	tests := []struct {
		name     string
		metadata metadata.MD
		wantNil  bool
	}{
		{
			name:     "nil metadata",
			metadata: nil,
			wantNil:  true,
		},
		{
			name:     "empty metadata",
			metadata: metadata.New(map[string]string{}),
			wantNil:  true,
		},
		{
			name: "valid trace metadata",
			metadata: metadata.New(map[string]string{
				sentry.SentryTraceHeader: "1234567890abcdef1234567890abcdef-1234567890abcdef-1",
			}),
			wantNil: false,
		},
		{
			name: "valid trace and baggage metadata",
			metadata: metadata.New(map[string]string{
				sentry.SentryTraceHeader:   "1234567890abcdef1234567890abcdef-1234567890abcdef-1",
				sentry.SentryBaggageHeader: "sentry-trace_id=1234567890abcdef1234567890abcdef",
			}),
			wantNil: false,
		},
		{
			name: "empty trace header",
			metadata: metadata.New(map[string]string{
				sentry.SentryTraceHeader: "",
			}),
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ContinueFromGrpcMetadata(tt.metadata)

			if tt.wantNil && result != nil {
				t.Errorf("Expected nil result, got %v", result)
			}
			if !tt.wantNil && result == nil {
				t.Errorf("Expected non-nil result, got nil")
			}
		})
	}
}

func TestToSpanStatus(t *testing.T) {
	tests := []struct {
		name     string
		code     codes.Code
		expected sentry.SpanStatus
	}{
		{"OK", codes.OK, sentry.SpanStatusOK},
		{"Canceled", codes.Canceled, sentry.SpanStatusCanceled},
		{"Unknown", codes.Unknown, sentry.SpanStatusUnknown},
		{"InvalidArgument", codes.InvalidArgument, sentry.SpanStatusInvalidArgument},
		{"DeadlineExceeded", codes.DeadlineExceeded, sentry.SpanStatusDeadlineExceeded},
		{"NotFound", codes.NotFound, sentry.SpanStatusNotFound},
		{"AlreadyExists", codes.AlreadyExists, sentry.SpanStatusAlreadyExists},
		{"PermissionDenied", codes.PermissionDenied, sentry.SpanStatusPermissionDenied},
		{"ResourceExhausted", codes.ResourceExhausted, sentry.SpanStatusResourceExhausted},
		{"FailedPrecondition", codes.FailedPrecondition, sentry.SpanStatusFailedPrecondition},
		{"Aborted", codes.Aborted, sentry.SpanStatusAborted},
		{"OutOfRange", codes.OutOfRange, sentry.SpanStatusOutOfRange},
		{"Unimplemented", codes.Unimplemented, sentry.SpanStatusUnimplemented},
		{"Internal", codes.Internal, sentry.SpanStatusInternalError},
		{"Unavailable", codes.Unavailable, sentry.SpanStatusUnavailable},
		{"DataLoss", codes.DataLoss, sentry.SpanStatusDataLoss},
		{"Unauthenticated", codes.Unauthenticated, sentry.SpanStatusUnauthenticated},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toSpanStatus(tt.code)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestUpdateFromSentryTrace(t *testing.T) {
	tests := []struct {
		name   string
		header []byte
	}{
		{
			name:   "valid trace header",
			header: []byte("1234567890abcdef1234567890abcdef-1234567890abcdef-1"),
		},
		{
			name:   "valid trace header without sampling",
			header: []byte("1234567890abcdef1234567890abcdef-1234567890abcdef"),
		},
		{
			name:   "invalid trace header",
			header: []byte("invalid-header"),
		},
		{
			name:   "empty header",
			header: []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use a valid span for testing
			span := sentry.StartSpan(context.Background(), "test-span")
			updateFromSentryTrace(span, tt.header)
			span.Finish()
		})
	}
}
