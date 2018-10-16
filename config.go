package ggpool

import "time"

// Config is a pool configuration
type Config struct {
	//Pool capacity.
	Capacity int

	//Pool keeps this number of objects ready for use.
	//If Item is destroyed and pool length less than MinCapacity then new Item will be created automatically.
	MinCapacity int

	//Duration of pool Item lifetime.
	//When the object lifetime expires method Object.Destroy() is called.
	ItemLifetime time.Duration

	//Item lifetime check period.
	//This means how often pool will check that the object lifetime is expired.
	ItemLifetimeCheckPeriod time.Duration

	//The timeout period of obtaining a item from the pool.
	//If the timeout is exceeded the pool will return Temporary error.
	Timeout time.Duration

	//Factory of pool Objects.
	//The Objects are wrapped by Item and can be obtained by Item.GetObject().
	Factory Creator
}
