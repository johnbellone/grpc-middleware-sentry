package grpc_sentry

import(
	"context"

	"github.com/getsentry/sentry-go"

	"google.golang.org/grpc"
)

func UnaryServerInterceptor(opts ...ServerOption) grpc.UnaryServerInterceptor {
	o := evaluateServerOptions(opts)
	return func(ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {
		panicked := true
		hub := sentry.GetHubFromContext(ctx)
		if hub == nil {
			hub = sentry.CurrentHub().Clone()
			ctx = sentry.SetHubOnContext(ctx, hub)
		}

		defer func() {
			var err error
			if err := recover(); err != nil || panicked {
				hub.RecoverWithContext(ctx, err)
			}

			if o.Repanic {
				panic(err)
			}
		}()

		resp, err := handler(ctx, req)
		panicked = false

		if o.ReportOn(err) {
			hub.CaptureException(err)
		}

		return resp, err
	}
}

func StreamServerInterceptor(opts ...ServerOption) grpc.StreamServerInterceptor {
	o := evaluateServerOptions(opts)
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		panicked := true
		ctx := ss.Context()
		hub := sentry.GetHubFromContext(ctx)
		if hub == nil {
			hub = sentry.CurrentHub().Clone()
			ctx = sentry.SetHubOnContext(ctx, hub)
		}

		defer func() {
			var err error
			if err := recover(); err != nil || panicked {
				hub.RecoverWithContext(ctx, err)
			}

			if o.Repanic {
				panic(err)
			}
		}()

		err := handler(srv, ss)
		panicked = false

		if err != nil && o.ReportOn(err) {
			hub.CaptureException(err)
		}

		return err
	}
}
