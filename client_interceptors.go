package grpc_sentry

import (
	"context"

	"github.com/getsentry/sentry-go"

	"google.golang.org/grpc"
)

func UnaryClientInterceptor(opts ...ClientOption) grpc.UnaryClientInterceptor {
	o := evaluateClientOptions(opts)
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

		err := invoker(ctx, method, req, reply, cc, callOpts...)

		if err != nil && o.ReportOn(err) {
			hub.CaptureException(err)
		}

		return err
	}
}

func StreamClientInterceptor(opts ...ClientOption) grpc.StreamClientInterceptor {
	o := evaluateClientOptions(opts)
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

		clientStream, err := streamer(ctx, desc, cc, method, callOpts...)

		if err != nil && o.ReportOn(err) {
			hub.CaptureException(err)
		}

		return clientStream, err
	}
}
