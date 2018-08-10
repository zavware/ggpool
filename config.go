package ggpool

import "time"

type Config struct {
	Capacity            int
	MinCapacity         int
	Lifetime            time.Duration
	LifetimeCheckPeriod time.Duration
	Timeout             time.Duration
	Factory             Creator
}
