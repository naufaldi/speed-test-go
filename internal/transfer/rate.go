package transfer

import "time"

// EWMA implements Exponential Weighted Moving Average for rate smoothing
type EWMA struct {
	alpha       float64
	value       float64
	initialized bool
}

// NewEWMA creates a new EWMA rate calculator
func NewEWMA(alpha float64) *EWMA {
	return &EWMA{
		alpha: alpha,
	}
}

// Update adds a new observation and returns the smoothed rate
func (e *EWMA) Update(newValue float64) float64 {
	if !e.initialized {
		e.value = newValue
		e.initialized = true
		return e.value
	}

	e.value = e.alpha*newValue + (1-e.alpha)*e.value
	return e.value
}

// Value returns the current smoothed value
func (e *EWMA) Value() float64 {
	return e.value
}

// RateCalculator calculates transfer rates
type RateCalculator struct {
	ewma        *EWMA
	startTime   time.Time
	lastTime    time.Time
	startBytes  int64
	lastBytes   int64
	currentRate float64
}

// NewRateCalculator creates a new rate calculator
func NewRateCalculator() *RateCalculator {
	return &RateCalculator{
		ewma: NewEWMA(0.1), // 10% weight for new values
	}
}

// Start begins rate calculation
func (rc *RateCalculator) Start() {
	now := time.Now()
	rc.startTime = now
	rc.lastTime = now
	rc.startBytes = 0
	rc.lastBytes = 0
	rc.currentRate = 0
	rc.ewma = NewEWMA(0.1)
}

// Update updates the rate based on current bytes transferred
func (rc *RateCalculator) Update(bytes int64) {
	now := time.Now()
	elapsed := now.Sub(rc.lastTime)

	if elapsed == 0 {
		return
	}

	deltaBytes := bytes - rc.lastBytes
	rc.lastBytes = bytes
	rc.lastTime = now

	if deltaBytes <= 0 {
		return
	}

	rate := float64(deltaBytes) / elapsed.Seconds()
	rc.currentRate = rc.ewma.Update(rate)
}

// SetBytes sets the total bytes transferred
func (rc *RateCalculator) SetBytes(bytes int64) {
	rc.Update(bytes)
}

// Rate returns the current rate in bytes per second
func (rc *RateCalculator) Rate() float64 {
	return rc.currentRate
}

// Reset resets the rate calculator
func (rc *RateCalculator) Reset() {
	now := time.Now()
	rc.startTime = now
	rc.lastTime = now
	rc.startBytes = 0
	rc.lastBytes = 0
	rc.currentRate = 0
	rc.ewma = NewEWMA(0.1)
}
