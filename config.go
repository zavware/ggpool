package ggpool

import (
	"errors"
	"time"
)

// Config is a pool configuration
type Config struct {
	//Pool capacity.
	Capacity int

	//Pool keeps this number of objects ready for use.
	//If Item is destroyed and pool length less than MinCapacity then new Object will be created automatically.
	MinCapacity int

	//Duration of pool Object lifetime.
	//When the object lifetime expires method Object.Destroy() is called.
	ItemLifetime time.Duration

	//Item lifetime check period.
	//This means how often pool will check that the object lifetime is expired.
	ItemLifetimeCheckPeriod time.Duration

	//The timeout period of obtaining a item from the pool.
	//If the timeout is exceeded the pool will return Temporary error.
	Timeout time.Duration

	//Factory of pool Objects.
	Factory Creator
}

func (c Config) validate() error {
	var err error

	if c.Capacity < 1 {
		err = errors.New("pool capacity value must be more then 0")
	}

	if c.MinCapacity < 0 {
		err = errors.New("min pool capacity value must not be negative")
	}

	if c.Capacity < c.MinCapacity {
		err = errors.New("pool capacity value cannot be less then init capacity value")
	}

	//@TODO to add validation for ItemLifetime, ItemLifetimeCheckPeriod, Timeout

	return err
}
