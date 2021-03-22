package grpc_sentry

import(
	"context"

	"github.com/getsentry/sentry-go"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_tags "github.com/grpc-ecosystem/go-grpc-middleware/tags"

	"google.golang.org/grpc"
)

func recoverWithSentry(hub *sentry.Hub, ctx context.Context, o *serveroptions) {
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

func UnaryServerInterceptor(opts ...ServerOption) grpc.UnaryServerInterceptor {
	o := evaluateServerOptions(opts)
	return func(ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {

		hub := sentry.GetHubFromContext(ctx)
		if hub == nil {
			hub = sentry.CurrentHub().Clone()
			ctx = sentry.SetHubOnContext(ctx, hub)
		}

		span := sentry.StartSpan(ctx, "grpc.server")
		defer span.Finish()

		// TODO: Perhaps makes sense to use SetRequestBody instead?
		hub.Scope().SetExtra("requestBody", req)
		hub.Scope().SetTransaction(info.FullMethod)
		defer recoverWithSentry(hub, ctx, o)

		resp, err := handler(ctx, req)
		if err != nil && o.ReportOn(err) {
			tags := grpc_tags.Extract(ctx)
			for k, v := range tags.Values() {
				hub.Scope().SetTag(k, v.(string))
			}

			hub.CaptureException(err)
		}

		return resp, err
	}
}

func StreamServerInterceptor(opts ...ServerOption) grpc.StreamServerInterceptor {
	o := evaluateServerOptions(opts)
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

		span := sentry.StartSpan(ctx, "grpc.server")
		defer span.Finish()

		stream := grpc_middleware.WrapServerStream(ss)
		stream.WrappedContext = ctx

		hub.Scope().SetTransaction(info.FullMethod)
		defer recoverWithSentry(hub, ctx, o)

		err := handler(srv, stream)
		if err != nil && o.ReportOn(err) {
			tags := grpc_tags.Extract(ctx)
			for k, v := range tags.Values() {
				hub.Scope().SetTag(k, v.(string))
			}

			hub.CaptureException(err)
		}

		return err
	}
}
