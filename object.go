package ggpool

type Object interface {
	Destroy()
	IsActive() bool
}
