package grpc_sentry

import (
	"context"

	"github.com/getsentry/sentry-go"
	"google.golang.org/grpc/metadata"

	"google.golang.org/grpc"
)

func UnaryClientInterceptor(opts ...Option) grpc.UnaryClientInterceptor {
	o := newConfig(opts)
	return func(ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		callOpts ...grpc.CallOption) error {

		hub := sentry.GetHubFromContext(ctx)
		if hub == nil {
			hub = sentry.CurrentHub().Clone()
			ctx = sentry.SetHubOnContext(ctx, hub)
		}

		operationName := defaultClientOperationName
		if o.OperationNameOverride != "" {
			operationName = o.OperationNameOverride
		}

		span := sentry.StartSpan(ctx, operationName, sentry.WithDescription(method))
		span.SetData("grpc.request.method", method)
		ctx = span.Context()
		md, ok := metadata.FromOutgoingContext(ctx)
		if ok {
			md.Append(sentry.SentryTraceHeader, span.ToSentryTrace())
			md.Append(sentry.SentryBaggageHeader, span.ToBaggage())
		} else {
			md = metadata.Pairs(
				sentry.SentryTraceHeader, span.ToSentryTrace(),
				sentry.SentryBaggageHeader, span.ToBaggage(),
			)
		}
		ctx = metadata.NewOutgoingContext(ctx, md)
		defer span.Finish()

		err := invoker(ctx, method, req, reply, cc, callOpts...)

		if err != nil && o.ReportOn(err) {
			hub.CaptureException(err)
		}

		return err
	}
}

func StreamClientInterceptor(opts ...Option) grpc.StreamClientInterceptor {
	o := newConfig(opts)
	return func(ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		callOpts ...grpc.CallOption) (grpc.ClientStream, error) {

		hub := sentry.GetHubFromContext(ctx)
		if hub == nil {
			hub = sentry.CurrentHub().Clone()
			ctx = sentry.SetHubOnContext(ctx, hub)
		}

		operationName := defaultClientOperationName
		if o.OperationNameOverride != "" {
			operationName = o.OperationNameOverride
		}

		span := sentry.StartSpan(ctx, operationName, sentry.WithDescription(method))
		span.SetData("grpc.request.method", method)
		ctx = span.Context()
		md, ok := metadata.FromOutgoingContext(ctx)
		if ok {
			md.Append(sentry.SentryTraceHeader, span.ToSentryTrace())
			md.Append(sentry.SentryBaggageHeader, span.ToBaggage())
		} else {
			md = metadata.Pairs(sentry.SentryTraceHeader, span.ToSentryTrace())
			md = metadata.Pairs(sentry.SentryBaggageHeader, span.ToBaggage())
		}
		ctx = metadata.NewOutgoingContext(ctx, md)
		defer span.Finish()

		clientStream, err := streamer(ctx, desc, cc, method, callOpts...)

		if err != nil && o.ReportOn(err) {
			hub.CaptureException(err)
		}

		return clientStream, err
	}
}
