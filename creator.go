package ggpool

type Creator interface {
	Create() (Object, error)
}
