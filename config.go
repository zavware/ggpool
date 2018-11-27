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
	//When the object lifetime expires method Object.Destroy() is called. Can be 0 - in this case item doesn't have lifetime limitation.
	ItemLifetime time.Duration

	//Item lifetime check period.
	//This means how often pool will check that the object lifetime is expired. If ItemLifetime is 0 then this setting is ignored
	ItemLifetimeCheckPeriod time.Duration

	//The timeout period of obtaining a item from the pool (Pool.Get()).
	//If the timeout is exceeded the pool will return TimeoutError error.
	Timeout time.Duration

	//Factory of pool Objects.
	Factory Creator
}

func (c Config) validate() error {

	if c.Capacity < 1 {
		return errors.New("pool capacity value must be more then 0")
	}

	if c.MinCapacity < 0 {
		return errors.New("min pool capacity value must not be negative")
	}

	if c.Capacity < c.MinCapacity {
		return errors.New("pool capacity value cannot be less then init capacity value")
	}

	if c.ItemLifetimeCheckPeriod == 0 && c.ItemLifetime > 0 {
		return errors.New("please specify ItemLifetimeCheckPeriod")
	}

	return nil
}
