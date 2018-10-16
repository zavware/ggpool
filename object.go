package ggpool

//Object is interface of Item object. Object that is created by Factory (see Config) must implement this interface
type Object interface {
	//Destroy is called when ItemLifetime is exceeded
	Destroy()
	//IsActive is called each time of ItemLifetimeCheckPeriod
	IsActive() bool
}
