# grpc-middleware-sentry

## Usage

``` go
package main

import (
    "github.com/getsentry/sentry-go"

    grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
    grpc_sentry "github.com/johnbellone/grpc-middleware-sentry"

    "google.golang.org/grpc"
)

const (
	Version = "0.1.0"
	SentryDsn = "https://897a3ef46125472da3ab8766deb302fe7fc7ade3@sentry.io/42"
)

func main() {
	err = sentry.Init(sentry.ClientOptions{
		Dsn: SentryDsn,
		Debug: false,
		Environment: "development",
		Release: Version,
		IgnoreErrors: []string{},
	})
	defer sentry.Flush(2 * time.Second)
	if err != nil {
		logger.Fatal(err.Error())
	}

	s := grpc.NewServer(
		grpc.Creds(creds),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_sentry.StreamServerInterceptor(),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_sentry.UnaryServerInterceptor(),
		)),
	)
}
```
