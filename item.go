package ggpool

import (
	"time"
)

type Item struct {
	object       *Object
	pool         *Pool
	releasedTime time.Time
}

func (i *Item) GetObject() *Object {
	return i.object
}

func (i *Item) Release() {
	i.releasedTime = time.Now().UTC()
	i.pool.items <- i
}

func (i *Item) Destroy() {
	(*i.GetObject()).Destroy()
}

func (i *Item) isActive() bool {
	expireTime := time.Now().Local().Add(i.pool.config.ItemLifetime)
	return i.releasedTime.Before(expireTime) && (*i.GetObject()).IsActive()
}
