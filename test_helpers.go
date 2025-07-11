// SPDX-License-Identifier: Apache-2.0
package grpc_sentry

import (
	"testing"
	"time"

	"github.com/getsentry/sentry-go"
)

// initSentryForTest initializes Sentry for testing with a mock transport
func initSentryForTest(t *testing.T) {
	t.Helper()

	// Initialize Sentry with a mock transport that discards events
	err := sentry.Init(sentry.ClientOptions{
		Dsn:        "https://test@test.ingest.sentry.io/123",
		Transport:  &sentry.HTTPSyncTransport{},
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event { return nil },
		Debug:      false,
	})
	if err != nil {
		t.Fatalf("Failed to initialize Sentry: %v", err)
	}

	// Ensure Sentry is flushed after tests
	t.Cleanup(func() {
		sentry.Flush(2 * time.Second)
	})
}

// initSentryForTestWithHub initializes Sentry and returns a hub for testing
func initSentryForTestWithHub(t *testing.T) *sentry.Hub {
	t.Helper()

	initSentryForTest(t)

	// Create a new hub for testing
	hub := sentry.CurrentHub().Clone()
	return hub
}
