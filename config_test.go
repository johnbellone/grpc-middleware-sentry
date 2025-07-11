// SPDX-License-Identifier: Apache-2.0
package grpc_sentry

import (
	"testing"
	"time"
)

func TestNewConfig_DefaultOptions(t *testing.T) {
	config := newConfig([]Option{})

	if config.Repanic != false {
		t.Errorf("Expected Repanic to be false, got %v", config.Repanic)
	}
	if config.WaitForDelivery != false {
		t.Errorf("Expected WaitForDelivery to be false, got %v", config.WaitForDelivery)
	}
	if config.Timeout != 1*time.Second {
		t.Errorf("Expected Timeout to be 1s, got %v", config.Timeout)
	}
	if config.OperationNameOverride != "" {
		t.Errorf("Expected OperationNameOverride to be empty, got %v", config.OperationNameOverride)
	}
	if config.CaptureRequestBody != true {
		t.Errorf("Expected CaptureRequestBody to be true, got %v", config.CaptureRequestBody)
	}
	if config.ReportOn == nil {
		t.Error("Expected ReportOn to be set, got nil")
	}
}

func TestNewConfig_WithRepanicOption(t *testing.T) {
	config := newConfig([]Option{WithRepanicOption(true)})

	if config.Repanic != true {
		t.Errorf("Expected Repanic to be true, got %v", config.Repanic)
	}
}

func TestNewConfig_WithWaitForDelivery(t *testing.T) {
	config := newConfig([]Option{WithWaitForDelivery(true)})

	if config.WaitForDelivery != true {
		t.Errorf("Expected WaitForDelivery to be true, got %v", config.WaitForDelivery)
	}
}

func TestNewConfig_WithTimeout(t *testing.T) {
	timeout := 5 * time.Second
	config := newConfig([]Option{WithTimeout(timeout)})

	if config.Timeout != timeout {
		t.Errorf("Expected Timeout to be %v, got %v", timeout, config.Timeout)
	}
}

func TestNewConfig_WithOperationNameOverride(t *testing.T) {
	operationName := "custom.operation"
	config := newConfig([]Option{WithOperationNameOverride(operationName)})

	if config.OperationNameOverride != operationName {
		t.Errorf("Expected OperationNameOverride to be %s, got %s", operationName, config.OperationNameOverride)
	}
}

func TestNewConfig_WithCaptureRequestBody(t *testing.T) {
	config := newConfig([]Option{WithCaptureRequestBody(false)})

	if config.CaptureRequestBody != false {
		t.Errorf("Expected CaptureRequestBody to be false, got %v", config.CaptureRequestBody)
	}
}

func TestNewConfig_WithReportOn(t *testing.T) {
	customReporter := func(error) bool { return false }
	config := newConfig([]Option{WithReportOn(customReporter)})

	if config.ReportOn == nil {
		t.Error("Expected ReportOn to be set, got nil")
	}
}

func TestNewConfig_MultipleOptions(t *testing.T) {
	config := newConfig([]Option{
		WithRepanicOption(true),
		WithWaitForDelivery(true),
		WithTimeout(10 * time.Second),
		WithOperationNameOverride("test.operation"),
		WithCaptureRequestBody(false),
	})

	if config.Repanic != true {
		t.Errorf("Expected Repanic to be true, got %v", config.Repanic)
	}
	if config.WaitForDelivery != true {
		t.Errorf("Expected WaitForDelivery to be true, got %v", config.WaitForDelivery)
	}
	if config.Timeout != 10*time.Second {
		t.Errorf("Expected Timeout to be 10s, got %v", config.Timeout)
	}
	if config.OperationNameOverride != "test.operation" {
		t.Errorf("Expected OperationNameOverride to be 'test.operation', got %s", config.OperationNameOverride)
	}
	if config.CaptureRequestBody != false {
		t.Errorf("Expected CaptureRequestBody to be false, got %v", config.CaptureRequestBody)
	}
}

func TestNewConfig_Validation(t *testing.T) {
	// Test with negative timeout - should be clamped to minimum
	config := newConfig([]Option{WithTimeout(-1 * time.Second)})
	if config.Timeout <= 0 {
		t.Errorf("Expected Timeout to be positive, got %v", config.Timeout)
	}

	// Test with zero timeout - should be clamped to minimum
	config = newConfig([]Option{WithTimeout(0)})
	if config.Timeout <= 0 {
		t.Errorf("Expected Timeout to be positive, got %v", config.Timeout)
	}

	// Test with nil ReportOn - should be set to ReportAlways
	config = newConfig([]Option{WithReportOn(nil)})
	if config.ReportOn == nil {
		t.Error("Expected ReportOn to be set to ReportAlways, got nil")
	}
}

func TestOption_Apply(t *testing.T) {
	// Test RepanicOption
	repanicOpt := &repanicOption{Repanic: true}
	config := &options{}
	repanicOpt.Apply(config)
	if config.Repanic != true {
		t.Errorf("Expected Repanic to be true after applying option, got %v", config.Repanic)
	}

	// Test WaitForDeliveryOption
	waitOpt := &waitForDeliveryOption{WaitForDelivery: true}
	config = &options{}
	waitOpt.Apply(config)
	if config.WaitForDelivery != true {
		t.Errorf("Expected WaitForDelivery to be true after applying option, got %v", config.WaitForDelivery)
	}

	// Test TimeoutOption
	timeout := 5 * time.Second
	timeoutOpt := &timeoutOption{Timeout: timeout}
	config = &options{}
	timeoutOpt.Apply(config)
	if config.Timeout != timeout {
		t.Errorf("Expected Timeout to be %v after applying option, got %v", timeout, config.Timeout)
	}

	// Test ReportOnOption
	customReporter := func(error) bool { return false }
	reportOpt := &reportOnOption{ReportOn: customReporter}
	config = &options{}
	reportOpt.Apply(config)
	if config.ReportOn == nil {
		t.Error("Expected ReportOn to be set after applying option, got nil")
	}

	// Test OperationNameOverride
	operationName := "test.operation"
	operationOpt := &operationNameOverride{OperationNameOverride: operationName}
	config = &options{}
	operationOpt.Apply(config)
	if config.OperationNameOverride != operationName {
		t.Errorf("Expected OperationNameOverride to be %s after applying option, got %s", operationName, config.OperationNameOverride)
	}

	// Test CaptureRequestBodyOption
	captureOpt := &captureRequestBodyOption{CaptureRequestBody: false}
	config = &options{}
	captureOpt.Apply(config)
	if config.CaptureRequestBody != false {
		t.Errorf("Expected CaptureRequestBody to be false after applying option, got %v", config.CaptureRequestBody)
	}
}
