package ggpool

//Object is interface which ggpool.Config.Factory objects must implement
type Object interface {
	Destroy()
	IsActive() bool
}
