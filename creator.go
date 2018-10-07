package ggpool

type Creator interface {
	Create() (interface{}, error)
}
