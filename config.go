// SPDX-License-Identifier: Apache-2.0
package grpc_sentry

import "time"

type Option interface {
	Apply(*options)
}

// newConfig returns a config configured with all the passed Options.
func newConfig(opts []Option) *options {
	optsCopy := *defaultOptions
	c := &optsCopy
	for _, o := range opts {
		o.Apply(c)
	}
	return c
}

type repanicOption struct {
	Repanic bool
}

func (r *repanicOption) Apply(o *options) {
	o.Repanic = r.Repanic
}

func WithRepanicOption(b bool) Option {
	return &repanicOption{Repanic: b}
}

type waitForDeliveryOption struct {
	WaitForDelivery bool
}

func (w *waitForDeliveryOption) Apply(o *options) {
	o.WaitForDelivery = w.WaitForDelivery
}

func WithWaitForDelivery(b bool) Option {
	return &waitForDeliveryOption{WaitForDelivery: b}
}

type timeoutOption struct {
	Timeout time.Duration
}

func (t *timeoutOption) Apply(o *options) {
	o.Timeout = t.Timeout
}

func WithTimeout(t time.Duration) Option {
	return &timeoutOption{Timeout: t}
}

type reporter func(error) bool

type reportOnOption struct {
	ReportOn reporter
}

func (r *reportOnOption) Apply(o *options) {
	o.ReportOn = r.ReportOn
}

func WithReportOn(r reporter) Option {
	return &reportOnOption{ReportOn: r}
}

type operationNameOverride struct {
	OperationNameOverride string
}

func (r *operationNameOverride) Apply(o *options) {
	o.OperationNameOverride = r.OperationNameOverride
}

func WithOperationNameOverride(s string) Option {
	return &operationNameOverride{OperationNameOverride: s}
}

type captureRequestBodyOption struct {
	CaptureRequestBody bool
}

func (c *captureRequestBodyOption) Apply(o *options) {
	o.CaptureRequestBody = c.CaptureRequestBody
}

func WithCaptureRequestBody(b bool) Option {
	return &captureRequestBodyOption{CaptureRequestBody: b}
}
