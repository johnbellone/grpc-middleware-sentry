// SPDX-License-Identifier: Apache-2.0
package grpc_sentry

import (
	"context"
	"encoding/hex"
	"regexp"

	"github.com/getsentry/sentry-go"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2"
	grpc_logging "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc"
)

func recoverWithSentry(hub *sentry.Hub, ctx context.Context, o *options) {
	if err := recover(); err != nil {
		eventID := hub.RecoverWithContext(ctx, err)
		if eventID != nil && o.WaitForDelivery {
			hub.Flush(o.Timeout)
		}

		if o.Repanic {
			panic(err)
		}
	}
}

func UnaryServerInterceptor(opts ...Option) grpc.UnaryServerInterceptor {
	o := newConfig(opts)
	return func(ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {

		hub := sentry.GetHubFromContext(ctx)
		if hub == nil {
			hub = sentry.CurrentHub().Clone()
			ctx = sentry.SetHubOnContext(ctx, hub)
		}

		operationName := defaultServerOperationName
		if o.OperationNameOverride != "" {
			operationName = o.OperationNameOverride
		}

		md, _ := metadata.FromIncomingContext(ctx) // nil check in ContinueFromGrpcMetadata

		// Use the FullMethod as transaction name and as description. This way the FullMethod will show up under
		// the span, and under the transaction.
		tx := sentry.StartTransaction(
			ctx,
			info.FullMethod,
			sentry.WithOpName(operationName),
			sentry.WithDescription(info.FullMethod),
			sentry.WithTransactionSource(sentry.SourceURL),
			ContinueFromGrpcMetadata(md),
		)
		tx.SetData("grpc.request.method", info.FullMethod)
		ctx = tx.Context()
		defer tx.Finish()

		if o.CaptureRequestBody {
			// TODO: Perhaps makes sense to use SetRequestBody instead?
			hub.Scope().SetExtra("requestBody", req)
		}
		defer recoverWithSentry(hub, ctx, o)

		resp, err := handler(ctx, req)
		if err != nil && o.ReportOn(err) {
			for k, v := range prepareLoggingFields(ctx) {
				hub.Scope().SetTag(k, v)
			}

			hub.CaptureException(err)

			// Always sample when an error has occurred.
			tx.Sampled = sentry.SampledTrue
		}
		tx.Status = toSpanStatus(status.Code(err))

		return resp, err
	}
}

func StreamServerInterceptor(opts ...Option) grpc.StreamServerInterceptor {
	o := newConfig(opts)
	return func(srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler) error {

		ctx := ss.Context()
		hub := sentry.GetHubFromContext(ctx)
		if hub == nil {
			hub = sentry.CurrentHub().Clone()
			ctx = sentry.SetHubOnContext(ctx, hub)
		}

		operationName := defaultServerOperationName
		if o.OperationNameOverride != "" {
			operationName = o.OperationNameOverride
		}

		md, _ := metadata.FromIncomingContext(ctx) // nil check in ContinueFromGrpcMetadata

		// Use the FullMethod as transaction name and as description. This way the FullMethod will show up under
		// the span, and under the transaction.
		tx := sentry.StartTransaction(
			ctx,
			info.FullMethod,
			sentry.WithOpName(operationName),
			sentry.WithDescription(info.FullMethod),
			sentry.WithTransactionSource(sentry.SourceURL),
			ContinueFromGrpcMetadata(md),
		)
		tx.SetData("grpc.request.method", info.FullMethod)
		ctx = tx.Context()
		defer tx.Finish()

		stream := grpc_middleware.WrapServerStream(ss)
		stream.WrappedContext = ctx

		defer recoverWithSentry(hub, ctx, o)

		err := handler(srv, stream)
		if err != nil && o.ReportOn(err) {
			for k, v := range prepareLoggingFields(ctx) {
				hub.Scope().SetTag(k, v)
			}

			hub.CaptureException(err)

			// Always sample when an error has occurred.
			tx.Sampled = sentry.SampledTrue
		}
		tx.Status = toSpanStatus(status.Code(err))

		return err
	}
}

// ContinueFromGrpcMetadata returns a span option that updates the span to continue
// an existing trace. If it cannot detect an existing trace in the request, the
// span will be left unchanged.
func ContinueFromGrpcMetadata(md metadata.MD) sentry.SpanOption {
	if md == nil {
		return nil
	}

	var trace, baggage string
	if traceMetadata, ok := md[sentry.SentryTraceHeader]; ok && len(traceMetadata) > 0 {
		trace = traceMetadata[0]
	}
	if baggageMetadata, ok := md[sentry.SentryBaggageHeader]; ok && len(baggageMetadata) > 0 {
		baggage = baggageMetadata[0]
	}

	// Only return span option if we have valid trace data
	if trace != "" {
		return sentry.ContinueFromHeaders(trace, baggage)
	}
	return nil
}

// Re-export of functions from tracing.go of sentry-go
var sentryTracePattern = regexp.MustCompile(`^([[:xdigit:]]{32})-([[:xdigit:]]{16})(?:-([01]))?$`)

func updateFromSentryTrace(s *sentry.Span, header []byte) {
	m := sentryTracePattern.FindSubmatch(header)
	if m == nil {
		// no match
		return
	}
	if _, err := hex.Decode(s.TraceID[:], m[1]); err != nil {
		// Log error for debugging but don't fail the operation
		sentry.GetHubFromContext(context.Background()).CaptureException(err)
	}
	if _, err := hex.Decode(s.ParentSpanID[:], m[2]); err != nil {
		// Log error for debugging but don't fail the operation
		sentry.GetHubFromContext(context.Background()).CaptureException(err)
	}
	if len(m[3]) != 0 {
		switch m[3][0] {
		case '0':
			s.Sampled = sentry.SampledFalse
		case '1':
			s.Sampled = sentry.SampledTrue
		}
	}
}

func toSpanStatus(code codes.Code) sentry.SpanStatus {
	switch code {
	case codes.OK:
		return sentry.SpanStatusOK
	case codes.Canceled:
		return sentry.SpanStatusCanceled
	case codes.Unknown:
		return sentry.SpanStatusUnknown
	case codes.InvalidArgument:
		return sentry.SpanStatusInvalidArgument
	case codes.DeadlineExceeded:
		return sentry.SpanStatusDeadlineExceeded
	case codes.NotFound:
		return sentry.SpanStatusNotFound
	case codes.AlreadyExists:
		return sentry.SpanStatusAlreadyExists
	case codes.PermissionDenied:
		return sentry.SpanStatusPermissionDenied
	case codes.ResourceExhausted:
		return sentry.SpanStatusResourceExhausted
	case codes.FailedPrecondition:
		return sentry.SpanStatusFailedPrecondition
	case codes.Aborted:
		return sentry.SpanStatusAborted
	case codes.OutOfRange:
		return sentry.SpanStatusOutOfRange
	case codes.Unimplemented:
		return sentry.SpanStatusUnimplemented
	case codes.Internal:
		return sentry.SpanStatusInternalError
	case codes.Unavailable:
		return sentry.SpanStatusUnavailable
	case codes.DataLoss:
		return sentry.SpanStatusDataLoss
	case codes.Unauthenticated:
		return sentry.SpanStatusUnauthenticated
	default:
		return sentry.SpanStatusUndefined
	}
}

func prepareLoggingFields(ctx context.Context) map[string]string {
	ret := make(map[string]string)
	fields := grpc_logging.ExtractFields(ctx)
	var label string
	for k, v := range fields {
		if k%2 == 0 {
			label = v.(string)
		} else {
			ret[label] = v.(string)
		}
	}

	return ret
}
