package ggpool

import "time"

// Config is configuration object for pool
type Config struct {
	//Number of objects which can be located in the pool
	Capacity int
	//The number of objects that are guaranteed to be located in the pool
	MinCapacity int
	//Duration of item lifetime
	ItemLifetime time.Duration
	//Item lifetime check period
	ItemLifetimeCheckPeriod time.Duration
	//The timeout period of obtaining a item from the pool
	Timeout time.Duration
	//Item object factory which implements Creator interface
	Factory Creator
}
