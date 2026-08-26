package cpa

import (
	"fmt"
	"time"
)

const (
	RiskNone  = "none"
	RiskWatch = "watch"
	RiskAlert = "alert"
)

type Limits struct {
	DCPAKm     float64
	TCPAWindow time.Duration
}

func DefaultLimits() Limits {
	return Limits{
		DCPAKm:     0.5,
		TCPAWindow: 15 * time.Minute,
	}
}

func (r Result) TCPAFrom(now time.Time) time.Duration {
	return r.At.Sub(now)
}

func Classify(r Result, now time.Time, lim Limits) (string, error) {
	if r.DCPAKm < 0 {
		return "", fmt.Errorf("cpa: DCPA cannot be negative")
	}
	if lim.DCPAKm <= 0 {
		return "", fmt.Errorf("cpa: DCPA limit must be positive")
	}
	if lim.TCPAWindow <= 0 {
		return "", fmt.Errorf("cpa: TCPA window must be positive")
	}
	if r.DCPAKm > lim.DCPAKm {
		return RiskNone, nil
	}
	tcpa := r.TCPAFrom(now)
	if tcpa < 0 {
		return RiskNone, nil
	}
	if tcpa > lim.TCPAWindow {
		return RiskWatch, nil
	}
	return RiskAlert, nil
}

func StricterThan(a, b string) bool {
	rank := map[string]int{RiskNone: 0, RiskWatch: 1, RiskAlert: 2}
	return rank[a] > rank[b]
}
