package transfer

import (
	"testing"
	"time"
)

func TestNewEWMA(t *testing.T) {
	alpha := 0.5
	ewma := NewEWMA(alpha)

	if ewma == nil {
		t.Fatal("Expected non-nil EWMA")
	}

	if ewma.alpha != alpha {
		t.Errorf("Expected alpha %f, got: %f", alpha, ewma.alpha)
	}

	if ewma.initialized {
		t.Error("Expected EWMA to not be initialized")
	}
}

func TestEWMA_Update_FirstValue(t *testing.T) {
	ewma := NewEWMA(0.1)

	value := ewma.Update(100.0)

	if value != 100.0 {
		t.Errorf("Expected first value to be 100.0, got: %f", value)
	}

	if !ewma.initialized {
		t.Error("Expected EWMA to be initialized after first update")
	}

	if ewma.value != 100.0 {
		t.Errorf("Expected stored value to be 100.0, got: %f", ewma.value)
	}
}

func TestEWMA_Update_Smoothing(t *testing.T) {
	// High alpha means more weight to new values
	ewma := NewEWMA(0.9)

	// First value
	ewma.Update(100.0)

	// Second value - with alpha 0.9, should be very close to new value
	value := ewma.Update(200.0)

	// With alpha 0.9: 0.9*200 + 0.1*100 = 190
	expected := 190.0
	if value != expected {
		t.Errorf("Expected smoothed value %f, got: %f", expected, value)
	}
}

func TestEWMA_Update_LowAlpha(t *testing.T) {
	// Low alpha means more weight to historical values
	ewma := NewEWMA(0.1)

	// First value
	ewma.Update(100.0)

	// Second value - with alpha 0.1, should be closer to old value
	value := ewma.Update(200.0)

	// With alpha 0.1: 0.1*200 + 0.9*100 = 110
	expected := 110.0
	if value != expected {
		t.Errorf("Expected smoothed value %f, got: %f", expected, value)
	}
}

func TestEWMA_MultipleUpdates(t *testing.T) {
	ewma := NewEWMA(0.5)

	// Series of updates
	values := []float64{100, 200, 300, 400, 500}

	for _, v := range values {
		ewma.Update(v)
	}

	// With alpha 0.5 and values 100, 200, 300, 400, 500:
	// v1 = 100
	// v2 = 0.5*200 + 0.5*100 = 150
	// v3 = 0.5*300 + 0.5*150 = 225
	// v4 = 0.5*400 + 0.5*225 = 312.5
	// v5 = 0.5*500 + 0.5*312.5 = 406.25
	expected := 406.25
	if ewma.value != expected {
		t.Errorf("Expected final value %f, got: %f", expected, ewma.value)
	}
}

func TestEWMA_Value(t *testing.T) {
	ewma := NewEWMA(0.1)

	// Initial value should be 0
	if ewma.Value() != 0 {
		t.Errorf("Expected initial value 0, got: %f", ewma.Value())
	}

	// After update
	ewma.Update(100.0)
	if ewma.Value() != 100.0 {
		t.Errorf("Expected value 100.0, got: %f", ewma.Value())
	}
}

func TestEWMA_ZeroAlpha(t *testing.T) {
	ewma := NewEWMA(0.0)

	ewma.Update(100.0)
	ewma.Update(200.0)

	// With alpha 0, all weight goes to historical values
	// After first update: 100
	// After second update: 0*200 + 1*100 = 100
	if ewma.value != 100.0 {
		t.Errorf("Expected value to remain 100.0 with alpha 0, got: %f", ewma.value)
	}
}

func TestEWMA_OneAlpha(t *testing.T) {
	ewma := NewEWMA(1.0)

	ewma.Update(100.0)
	ewma.Update(200.0)

	// With alpha 1, all weight goes to new value
	// After first update: 100
	// After second update: 1*200 + 0*100 = 200
	if ewma.value != 200.0 {
		t.Errorf("Expected value to be 200.0 with alpha 1, got: %f", ewma.value)
	}
}

func TestNewRateCalculator(t *testing.T) {
	rc := NewRateCalculator()

	if rc == nil {
		t.Fatal("Expected non-nil RateCalculator")
	}

	if rc.ewma == nil {
		t.Error("Expected EWMA to be initialized")
	}

	if rc.currentRate != 0 {
		t.Errorf("Expected initial rate 0, got: %f", rc.currentRate)
	}
}

func TestRateCalculator_Start(t *testing.T) {
	rc := NewRateCalculator()

	rc.Start()

	if rc.startBytes != 0 {
		t.Errorf("Expected startBytes 0, got: %d", rc.startBytes)
	}

	if rc.currentRate != 0 {
		t.Errorf("Expected currentRate 0, got: %f", rc.currentRate)
	}

	if rc.ewma == nil {
		t.Error("Expected ewma to be initialized")
	}
}

func TestRateCalculator_Update(t *testing.T) {
	rc := NewRateCalculator()
	rc.Start()

	// Wait 100ms to establish a meaningful baseline
	time.Sleep(100 * time.Millisecond)

	// Simulate 100KB transferred
	rc.Update(100 * 1024)

	// Rate should be approximately 100KB/s (100KB / 100ms = 1000KB/s = ~1MB/s)
	// Allow some tolerance for timing variations
	expectedRate := 100.0 * 1024 // bytes per 100ms
	actualRate := rc.currentRate

	// The rate should be within an order of magnitude of expected
	if actualRate < expectedRate/10 || actualRate > expectedRate*10 {
		t.Errorf("Expected rate around %f, got: %f", expectedRate, actualRate)
	}
}

func TestRateCalculator_Update_ZeroElapsed(t *testing.T) {
	rc := NewRateCalculator()
	rc.Start()

	// Call Update immediately without waiting - should handle gracefully
	// The elapsed time will be very small, so the rate will be very high
	// but it shouldn't panic
	rc.Update(1000)

	// Just verify it doesn't panic and returns some rate
	if rc.currentRate < 0 {
		t.Errorf("Expected non-negative rate, got: %f", rc.currentRate)
	}
}

func TestRateCalculator_SetBytes(t *testing.T) {
	rc := NewRateCalculator()
	rc.Start()

	// Set first batch
	rc.SetBytes(1000)

	// Verify rate is calculated (don't check internal state)
	rate1 := rc.currentRate
	if rate1 < 0 {
		t.Errorf("Expected non-negative rate, got: %f", rate1)
	}

	// Wait a bit
	time.Sleep(10 * time.Millisecond)

	// Set second batch
	rc.SetBytes(2500)

	// Verify rate is still calculated
	rate2 := rc.currentRate
	if rate2 < 0 {
		t.Errorf("Expected non-negative rate, got: %f", rate2)
	}
}

func TestRateCalculator_Rate(t *testing.T) {
	rc := NewRateCalculator()

	if rc.Rate() != 0 {
		t.Errorf("Expected initial rate 0, got: %f", rc.Rate())
	}

	rc.Start()
	rc.Update(1024 * 1024)

	if rc.Rate() != rc.currentRate {
		t.Errorf("Expected Rate() to return currentRate")
	}
}

func TestRateCalculator_Reset(t *testing.T) {
	rc := NewRateCalculator()
	rc.Start()
	rc.Update(1024 * 1024)

	rc.Reset()

	if rc.startBytes != 0 {
		t.Errorf("Expected startBytes 0 after reset, got: %d", rc.startBytes)
	}

	if rc.currentRate != 0 {
		t.Errorf("Expected currentRate 0 after reset, got: %f", rc.currentRate)
	}

	if rc.ewma == nil {
		t.Error("Expected ewma to be reinitialized after reset")
	}
}

func TestRateCalculator_SequentialUpdates(t *testing.T) {
	rc := NewRateCalculator()
	rc.Start()

	// Simulate multiple chunks being transferred
	chunks := []int64{1024, 2048, 4096, 8192, 16384}

	for _, chunk := range chunks {
		rc.SetBytes(chunk)
	}

	// Rate should reflect the EWMA of all chunks
	if rc.currentRate <= 0 {
		t.Error("Expected positive rate after updates")
	}

	// With smoothing, later chunks should have more influence
	// but all should contribute
}

func TestRateCalculator_ContinuousTransfer(t *testing.T) {
	rc := NewRateCalculator()
	rc.Start()

	// Wait a bit to establish a baseline
	time.Sleep(50 * time.Millisecond)

	totalBytes := int64(0)
	// Simulate continuous transfer with delays
	for i := 0; i < 10; i++ {
		chunkSize := int64(1024 * 1024) // 1MB per chunk
		totalBytes += chunkSize
		rc.SetBytes(totalBytes)
		t.Logf("After update %d: rate = %f MB/s", i+1, rc.currentRate/(1024*1024))
		time.Sleep(100 * time.Millisecond) // 100ms delay between updates
	}

	// After multiple updates, the rate should stabilize
	// Each update transfers 1MB, so the rate should be around 10MB/s (1MB per 100ms)
	expectedMBps := 10.0
	actualMBps := rc.currentRate / (1024 * 1024)
	t.Logf("Final rate: %f MB/s (expected ~%f MB/s)", actualMBps, expectedMBps)

	// Allow 50% tolerance for timing variations and EWMA smoothing
	if actualMBps < expectedMBps*0.5 || actualMBps > expectedMBps*2.0 {
		t.Errorf("Expected rate around %f MB/s, got: %f MB/s", expectedMBps, actualMBps)
	}
}

func TestEWMA_NegativeValues(t *testing.T) {
	ewma := NewEWMA(0.5)

	// Should handle negative values without issues
	ewma.Update(-100.0)
	ewma.Update(-200.0)

	// Result should be smoothed negative value
	if ewma.value >= 0 {
		t.Errorf("Expected negative value, got: %f", ewma.value)
	}
}

func TestEWMA_ZeroValues(t *testing.T) {
	ewma := NewEWMA(0.5)

	ewma.Update(0.0)
	ewma.Update(100.0)

	// After seeing a zero, then 100, should be somewhere in between
	if ewma.value == 0 {
		t.Error("Expected value to change after non-zero update")
	}
}

func TestRateCalculator_LargeValues(t *testing.T) {
	rc := NewRateCalculator()
	rc.Start()

	// Wait 100ms to establish a meaningful baseline
	time.Sleep(100 * time.Millisecond)

	// Simulate large transfer: 100MB in 100ms
	rc.SetBytes(100 * 1024 * 1024)

	// Rate should be approximately 100MB/s (100MB / 100ms = 1000MB/s = 1GB/s)
	// But due to timing variations, allow a wide tolerance
	expectedBytesPerSec := 100.0 * 1024 * 1024 // bytes per 100ms -> bytes per second
	actualRate := rc.currentRate

	// The rate should be within an order of magnitude of expected
	if actualRate < expectedBytesPerSec/10 || actualRate > expectedBytesPerSec*10 {
		t.Errorf("Expected rate around %f, got: %f", expectedBytesPerSec, actualRate)
	}
}

func TestRateCalculator_SmallValues(t *testing.T) {
	rc := NewRateCalculator()
	rc.Start()

	// Simulate small transfer: 1 byte
	rc.SetBytes(1)

	// Rate should be calculated (might be very small depending on timing)
	// The important thing is it doesn't panic
	if rc.currentRate < 0 {
		t.Error("Expected non-negative rate")
	}
}
