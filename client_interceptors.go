package grpc_sentry

import (
	"context"
	"google.golang.org/grpc/metadata"

	"github.com/getsentry/sentry-go"

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
		
		hub.Scope().SetTransaction(method)

		span := sentry.StartSpan(ctx, "grpc.client")
		ctx = span.Context()
		md := metadata.Pairs("sentry-trace", span.ToSentryTrace())
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
		
		hub.Scope().SetTransaction(method)

		span := sentry.StartSpan(ctx, "grpc.client")
		ctx = span.Context()
		md := metadata.Pairs("sentry-trace", span.ToSentryTrace())
		ctx = metadata.NewOutgoingContext(ctx, md)
		defer span.Finish()

		clientStream, err := streamer(ctx, desc, cc, method, callOpts...)

		if err != nil && o.ReportOn(err) {
			hub.CaptureException(err)
		}

		return clientStream, err
	}
}
