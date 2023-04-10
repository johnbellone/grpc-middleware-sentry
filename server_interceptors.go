package grpc_sentry

import (
	"context"
	"encoding/hex"
	"regexp"

	"github.com/getsentry/sentry-go"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_tags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
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

		md, _ := metadata.FromIncomingContext(ctx) // nil check in ContinueFromGrpcMetadata
		span := sentry.StartSpan(ctx, "grpc.server", ContinueFromGrpcMetadata(md))
		ctx = span.Context()
		defer span.Finish()

		// TODO: Perhaps makes sense to use SetRequestBody instead?
		hub.Scope().SetExtra("requestBody", req)
		defer recoverWithSentry(hub, ctx, o)

		resp, err := handler(ctx, req)
		if err != nil && o.ReportOn(err) {
			tags := grpc_tags.Extract(ctx)
			for k, v := range tags.Values() {
				hub.Scope().SetTag(k, v.(string))
			}

			hub.CaptureException(err)
		}
		span.Status = toSpanStatus(status.Code(err))

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

		md, _ := metadata.FromIncomingContext(ctx) // nil check in ContinueFromGrpcMetadata
		span := sentry.StartSpan(ctx, "grpc.server", ContinueFromGrpcMetadata(md))
		ctx = span.Context()
		defer span.Finish()

		stream := grpc_middleware.WrapServerStream(ss)
		stream.WrappedContext = ctx

		defer recoverWithSentry(hub, ctx, o)

		err := handler(srv, stream)
		if err != nil && o.ReportOn(err) {
			tags := grpc_tags.Extract(ctx)
			for k, v := range tags.Values() {
				hub.Scope().SetTag(k, v.(string))
			}

			hub.CaptureException(err)
		}
		span.Status = toSpanStatus(status.Code(err))

		return err
	}
}

// ContinueFromGrpcMetadata returns a span option that updates the span to continue
// an existing trace. If it cannot detect an existing trace in the request, the
// span will be left unchanged.
func ContinueFromGrpcMetadata(md metadata.MD) sentry.SpanOption {
	return func(s *sentry.Span) {
		if md == nil {
			return
		}

		trace, ok := md["sentry-trace"]
		if !ok {
			return
		}
		if len(trace) != 1 {
			return
		}
		if trace[0] == "" {
			return
		}
		updateFromSentryTrace(s, []byte(trace[0]))
	}
}

// Re-export of functions from tracing.go of sentry-go
var sentryTracePattern = regexp.MustCompile(`^([[:xdigit:]]{32})-([[:xdigit:]]{16})(?:-([01]))?$`)

func updateFromSentryTrace(s *sentry.Span, header []byte) {
	m := sentryTracePattern.FindSubmatch(header)
	if m == nil {
		// no match
		return
	}
	_, _ = hex.Decode(s.TraceID[:], m[1])
	_, _ = hex.Decode(s.ParentSpanID[:], m[2])
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
