package ggpool

type Object interface {
	Destroy() (bool, error)
	IsActive() bool
}
