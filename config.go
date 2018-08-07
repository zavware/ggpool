package ggpool

import "time"

type Config struct {
	Capacity            int
	InitCapacity        int
	Lifetime            time.Duration
	LifetimeCheckPeriod time.Duration
	Timeout             time.Duration
	Factory             Creator
}
